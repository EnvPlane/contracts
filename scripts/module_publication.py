#!/usr/bin/env python3
"""Conservative release classification and immutable, main-only publication."""

import argparse
from dataclasses import dataclass
import fnmatch
import posixpath
import re
import shlex
import subprocess
import sys


VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\Z")
ROOT_DOCS = {"README.md", "CONTRIBUTING.md", "SECURITY.md", "TRADEMARKS.md", "CHANGELOG.md"}
PUBLICATION_METADATA = {
    ".github/workflows/publish-module.yaml",
    "scripts/module_publication.py",
    "scripts/tests/test_module_publication.py",
}


def git(*args):
    return subprocess.run(["git", *args], check=True, stdout=subprocess.PIPE).stdout


def commit(ref):
    return git("rev-parse", "--verify", "--end-of-options", ref + "^{commit}").decode().strip()


def latest_release():
    tags = git("tag", "--list").decode().splitlines()
    stable = [(tuple(map(int, VERSION.fullmatch(tag).groups())), tag)
              for tag in tags if VERSION.fullmatch(tag)]
    return max(stable)[1] if stable else None


@dataclass(frozen=True)
class Decision:
    publish: bool
    reason: str


def docs_path(path):
    return path in ROOT_DOCS or (path.startswith(("docs/", "issues/")) and path.endswith(".md"))


def regular_document(revision, path):
    # Deleted/new files are allowed; executable docs, symlinks and submodules are not.
    entry = git("ls-tree", "-z", revision, "--", ":(literal)" + path)
    return not entry or entry.startswith(b"100644 blob ")


def embedded_path_change(revision, paths):
    # Patterns are relative to the Go source directory. Unrelated OpenAPI embeds
    # must not turn every documentation/publisher change into a release.
    result = subprocess.run(["git", "grep", "-l", "-z", "-e", "go:embed", revision, "--", "*.go"],
                            stdout=subprocess.PIPE)
    if result.returncode not in (0, 1):
        raise RuntimeError("unable to inspect embedded module inputs")
    for entry in result.stdout.split(b"\0"):
        if not entry:
            continue
        source = entry.decode("utf-8", errors="surrogateescape").split(":", 1)[1]
        text = git("show", revision + ":" + source).decode("utf-8", errors="surrogateescape")
        for directive in re.findall(r"^\s*//go:embed\s+(.+)$", text, flags=re.MULTILINE):
            # Unusual Go quoting/escapes are conservative rather than guessed.
            if "`" in directive or "\\" in directive:
                return True
            try:
                patterns = shlex.split(directive)
            except ValueError:
                return True
            if not patterns:
                return True
            for pattern in patterns:
                pattern = pattern.removeprefix("all:")
                if (not pattern or "[" in pattern or pattern.startswith("/")
                        or ".." in pattern.split("/")):
                    return True
                for path in paths:
                    relative = posixpath.relpath(path, posixpath.dirname(source) or ".")
                    if relative.startswith("../"):
                        continue
                    parts = relative.split("/")
                    # Directory embeds include descendants. fnmatch's broader
                    # star matching can over-publish but never under-classify.
                    if any(fnmatch.fnmatchcase("/".join(parts[:i]), pattern)
                           for i in range(1, len(parts) + 1)):
                        return True
    return False


def classify(event, repository, ref, target, current_main=None):
    # No Git/file execution is needed for untrusted events or forks.
    if repository != "EnvPlane/contracts" or ref != "refs/heads/main" or event not in ("push", "workflow_dispatch"):
        return Decision(False, "untrusted-event-or-ref")
    target = commit(target)
    if commit("HEAD") != target:
        raise RuntimeError("checkout does not match event commit")
    if current_main is not None and commit(current_main) != target:
        return Decision(False, "superseded-main-run")
    latest = latest_release()
    if latest is None:
        return Decision(True, "initial-release")
    baseline = commit(latest)
    if baseline == target:
        return Decision(False, "already-published")
    if subprocess.run(["git", "merge-base", "--is-ancestor", baseline, target]).returncode != 0:
        raise RuntimeError("latest release is not an ancestor of the candidate; manual review required")
    if event == "workflow_dispatch":
        return Decision(True, "manual-main-dispatch")
    changed = git("diff", "--name-only", "--no-renames", "-z", baseline, target, "--").split(b"\0")
    paths = [path.decode("utf-8", errors="surrogateescape") for path in changed if path]
    if not paths:
        return Decision(False, "no-module-changes")
    for path in paths:
        if (not docs_path(path) and path not in PUBLICATION_METADATA) or not regular_document(baseline, path) or not regular_document(target, path):
            return Decision(True, "module-or-unknown-path-change")
    if embedded_path_change(baseline, paths) or embedded_path_change(target, paths):
        return Decision(True, "possible-embedded-input-change")
    reason = "publication-metadata-only" if any(path in PUBLICATION_METADATA for path in paths) else "documentation-only"
    return Decision(False, reason)


def next_version(latest):
    if latest is None:
        return "v0.1.0"
    match = VERSION.fullmatch(latest)
    if match is None:
        raise ValueError("expected canonical stable version")
    major, minor, patch = map(int, match.groups())
    return f"v{major}.{minor}.{patch + 1}"


def publish(event, repository, ref, target):
    if repository != "EnvPlane/contracts" or ref != "refs/heads/main" or event not in ("push", "workflow_dispatch"):
        raise RuntimeError("publication requires a trusted canonical main event")
    # Refresh again after tests. No force: a moved/replaced immutable tag fails fetch.
    git("fetch", "origin", "refs/heads/main:refs/remotes/origin/main", "--tags")
    if commit("refs/remotes/origin/main") != commit(target):
        return Decision(False, "superseded-main-run")
    decision = classify(event, repository, ref, target)
    if not decision.publish:
        return decision
    tag = next_version(latest_release())
    git("config", "user.name", "envplane automation")
    git("config", "user.email", "automation@envplane.dev")
    # git tag/push both reject collisions. Never delete, force, or move a tag.
    git("tag", "-a", tag, commit(target), "-m", f"Release {tag}")
    git("push", "origin", f"refs/tags/{tag}:refs/tags/{tag}")
    print(f"tag={tag}")
    return Decision(True, "published")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("classify", "publish"))
    for name in ("event", "repository", "ref", "commit"):
        parser.add_argument("--" + name, required=True)
    parser.add_argument("--current-main", help="freshly fetched main SHA for classification")
    args = parser.parse_args()
    operation = classify if args.action == "classify" else publish
    if args.action == "classify":
        decision = operation(args.event, args.repository, args.ref, args.commit, args.current_main)
    else:
        if args.current_main is not None:
            parser.error("publication refreshes main itself; --current-main is classification-only")
        decision = operation(args.event, args.repository, args.ref, args.commit)
    print(f"publish={str(decision.publish).lower()}")
    print(f"reason={decision.reason}")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError, subprocess.CalledProcessError) as error:
        print(f"publication refused: {error}", file=sys.stderr)
        sys.exit(1)
