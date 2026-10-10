"""Local Git fixtures only; no GitHub, live cluster, or public tag writes."""

import contextlib
import importlib.util
import io
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest import mock


SCRIPT = Path(__file__).resolve().parents[1] / "module_publication.py"
SPEC = importlib.util.spec_from_file_location("module_publication", SCRIPT)
publication = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = publication
SPEC.loader.exec_module(publication)


class PublicationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="contracts-publisher-")
        self.previous = Path.cwd()
        self.root = Path(self.temp.name) / "checkout"
        self.root.mkdir()
        os.chdir(self.root)
        self.git("init", "-q", "-b", "main")
        self.git("config", "user.name", "fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.write("README.md", "Initial documentation\n")
        self.save()

    def tearDown(self):
        os.chdir(self.previous)
        self.temp.cleanup()

    def git(self, *args):
        return subprocess.check_output(["git", *args], stderr=subprocess.PIPE).decode().strip()

    def write(self, path, content="metadata\n"):
        target = self.root / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)

    def save(self):
        self.git("add", "-A")
        self.git("commit", "-q", "--allow-empty", "-m", "fixture")
        return self.git("rev-parse", "HEAD")

    def baseline(self):
        self.git("tag", "-a", "v0.1.111", "-m", "fixture baseline")

    def decide(self, event="push", repository="envplane/contracts", ref="refs/heads/main", target=None):
        return publication.classify(event, repository, ref, target or self.git("rev-parse", "HEAD"))

    def remote(self):
        self.origin = Path(self.temp.name) / "origin.git"
        self.git("init", "-q", "--bare", str(self.origin))
        self.git("remote", "add", "origin", str(self.origin))
        self.git("push", "-q", "origin", "main", "--tags")

    def release(self, target=None, event="push"):
        with contextlib.redirect_stdout(io.StringIO()):
            return publication.publish(event, "envplane/contracts", "refs/heads/main",
                                       target or self.git("rev-parse", "HEAD"))

    def test_first_documentation_commit_publishes_initial_release(self):
        self.assertEqual(self.decide(), publication.Decision(True, "initial-release"))

    def test_docs_only_without_any_released_baseline_still_publishes(self):
        self.write("docs/intro.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_released_docs_only_skips_including_spaces_and_newlines(self):
        self.baseline()
        for path in ("README.md", "issues/2026-10-10/ticket.md", "docs/ADR/a.md", "docs/a b\nc.md"):
            self.write(path, "updated\n")
        self.save()
        self.assertEqual(self.decide(), publication.Decision(False, "documentation-only"))

    def test_docs_delete_and_rename_skip(self):
        self.write("docs/old.md")
        self.save()
        self.baseline()
        (self.root / "README.md").unlink()
        self.git("mv", "docs/old.md", "docs/new.md")
        self.save()
        self.assertFalse(self.decide().publish)

    def test_semantic_and_unknown_paths_publish(self):
        self.baseline()
        for path in ("domain/plan.go", "sdk/api.go", "go.mod", "go.sum", "openapi/openapi.json",
                     "schemas/plan.yaml", "docs/schema.json", "scripts/tool.sh", "LICENSE",
                     ".github/workflows/ci.yaml", "fixtures/input.md", "docs/example.go"):
            with self.subTest(path=path):
                self.write(path)
                self.save()
                self.assertTrue(self.decide().publish)
                (self.root / path).unlink()
                self.save()

    def test_mixed_docs_and_code_publish(self):
        self.baseline()
        self.write("docs/notes.md")
        self.write("domain/new.go", "package domain\n")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_exact_publisher_metadata_only_does_not_release(self):
        self.baseline()
        for path in publication.PUBLICATION_METADATA:
            self.write(path)
        self.write("issues/publisher.md")
        self.save()
        self.assertEqual(self.decide(), publication.Decision(False, "publication-metadata-only"))
        self.assertTrue(self.decide(event="workflow_dispatch").publish)

    def test_publisher_metadata_mixed_with_schema_still_releases(self):
        self.baseline()
        self.write("scripts/module_publication.py")
        self.write("openapi/openapi.json")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_unpublished_semantic_commit_is_not_hidden_by_later_docs_push(self):
        self.baseline()
        self.write("domain/new.go", "package domain\n")
        self.save()
        self.write("issues/ticket.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_semantic_deletion_and_rename_to_docs_publish(self):
        self.write("domain/input.json")
        self.save()
        self.baseline()
        (self.root / "docs").mkdir()
        self.git("mv", "domain/input.json", "docs/input.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_symlink_and_executable_docs_are_not_skipped(self):
        self.baseline()
        (self.root / "docs").mkdir()
        (self.root / "docs/link.md").symlink_to("../README.md")
        self.save()
        self.assertTrue(self.decide().publish)
        (self.root / "docs/link.md").unlink()
        self.write("docs/run.md")
        (self.root / "docs/run.md").chmod(0o755)
        self.save()
        self.assertTrue(self.decide().publish)

    def test_embedded_markdown_is_conservatively_published(self):
        self.write("embed.go", '//go:embed docs/*.md\n')
        self.save()
        self.baseline()
        self.write("docs/input.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_unrelated_openapi_embed_does_not_publish_docs_or_pipeline(self):
        self.write("openapi.go", "//go:embed openapi/openapi.json\n")
        self.save()
        self.baseline()
        self.write("docs/only.md")
        self.write("scripts/module_publication.py")
        self.save()
        self.assertFalse(self.decide().publish)

    def test_directory_and_quoted_embeds_publish_changed_docs(self):
        self.write("embed.go", '//go:embed "docs/a b.md" all:issues\n')
        self.save()
        self.baseline()
        self.write("issues/nested/ticket.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_uncertain_embed_quoting_is_not_guessed(self):
        self.write("embed.go", '//go:embed `docs/a b.md`\n')
        self.save()
        self.baseline()
        self.write("docs/only.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_uncertain_embed_character_class_is_conservative(self):
        self.write("embed.go", '//go:embed docs/[a-z].md\n')
        self.save()
        self.baseline()
        self.write("issues/only.md")
        self.save()
        self.assertTrue(self.decide().publish)

    def test_manual_docs_release_preserved_but_rerun_is_noop(self):
        self.baseline()
        self.assertFalse(self.decide(event="workflow_dispatch").publish)
        self.write("docs/manual.md")
        self.save()
        self.assertTrue(self.decide(event="workflow_dispatch").publish)
        self.assertFalse(self.decide().publish)

    def test_untrusted_events_refs_and_forks_do_not_read_checkout(self):
        for event, repo, ref in (("pull_request", "envplane/contracts", "refs/heads/main"),
                                 ("pull_request_target", "envplane/contracts", "refs/heads/main"),
                                 ("push", "attacker/contracts", "refs/heads/main"),
                                 ("workflow_dispatch", "envplane/contracts", "refs/heads/topic"),
                                 ("push", "envplane/contracts", "refs/tags/v0.1.112")):
            with self.subTest(event=event, repo=repo, ref=ref), mock.patch.object(publication, "git", side_effect=AssertionError("untrusted Git access")):
                self.assertFalse(publication.classify(event, repo, ref, "missing").publish)
                with self.assertRaises(RuntimeError):
                    publication.publish(event, repo, ref, "missing")

    def test_canonical_repository_matching_is_case_insensitive(self):
        self.assertTrue(self.decide(repository="envplane/contracts".upper()).publish)

    def test_checkout_identity_mismatch_refused(self):
        old = self.git("rev-parse", "HEAD")
        self.save()
        with self.assertRaises(RuntimeError):
            self.decide(target=old)

    def test_nonancestor_latest_release_refused(self):
        old = self.git("rev-parse", "HEAD")
        self.write("domain/input.go")
        self.save()
        self.baseline()
        self.git("checkout", "-q", "--detach", old)
        with self.assertRaises(RuntimeError):
            self.decide()

    def test_empty_commit_is_noop(self):
        self.baseline()
        self.save()
        self.assertFalse(self.decide().publish)

    def test_numeric_versions_ignore_prerelease_and_malformed_tags(self):
        for tag in ("v0.1.9", "v0.1.111", "v0.1.112-rc1", "v0.1.099", "v1.0.0-extra"):
            self.git("tag", tag)
        self.assertEqual(publication.latest_release(), "v0.1.111")
        self.assertEqual(publication.next_version("v0.1.111"), "v0.1.112")
        self.assertEqual(publication.next_version(None), "v0.1.0")
        with self.assertRaises(ValueError):
            publication.next_version("v0.1.112-rc1")

    def test_publish_semantic_tag_and_rerun_without_duplicate(self):
        self.baseline()
        self.write("domain/input.go")
        target = self.save()
        self.remote()
        self.assertTrue(self.release().publish)
        self.assertEqual(self.git("rev-parse", "v0.1.112^{commit}"), target)
        self.assertFalse(self.release().publish)
        self.assertEqual(publication.latest_release(), "v0.1.112")

    def test_docs_publish_does_not_allocate_a_remote_tag(self):
        self.baseline()
        self.write("issues/only.md")
        self.save()
        self.remote()
        self.assertFalse(self.release().publish)
        self.assertEqual(publication.latest_release(), "v0.1.111")

    def test_manual_docs_publish_allocates_once(self):
        self.baseline()
        self.write("docs/manual.md")
        target = self.save()
        self.remote()
        self.assertTrue(self.release(event="workflow_dispatch").publish)
        self.assertEqual(self.git("rev-parse", "v0.1.112^{commit}"), target)
        self.assertFalse(self.release(event="workflow_dispatch").publish)

    def test_superseded_main_run_is_noop(self):
        self.baseline()
        self.write("domain/input.go")
        old = self.save()
        self.remote()
        self.write("domain/new.go")
        self.save()
        self.git("push", "-q", "origin", "main")
        self.git("checkout", "-q", "--detach", old)
        self.assertEqual(self.release(target=old), publication.Decision(False, "superseded-main-run"))
        self.assertEqual(publication.latest_release(), "v0.1.111")

    def test_classification_skips_stale_main_even_after_new_release(self):
        self.baseline()
        self.write("domain/input.go")
        old = self.save()
        self.write("domain/new.go")
        newer = self.save()
        self.git("tag", "v0.1.112")
        self.git("checkout", "-q", "--detach", old)
        result = publication.classify("push", "envplane/contracts", "refs/heads/main", old, newer)
        self.assertEqual(result, publication.Decision(False, "superseded-main-run"))

    def test_fetch_failure_is_not_treated_as_missing_tag(self):
        self.baseline()
        self.write("domain/input.go")
        self.save()
        with self.assertRaises(subprocess.CalledProcessError):
            self.release()
        self.assertEqual(publication.latest_release(), "v0.1.111")

    def test_workflow_keeps_trust_concurrency_and_release_gates(self):
        workflow = (SCRIPT.parents[1] / ".github/workflows/publish-module.yaml").read_text()
        for required in ("github.repository == 'envplane/contracts'", "github.ref == 'refs/heads/main'",
                         "github.event_name == 'push'", "github.event_name == 'workflow_dispatch'",
                         "group: contracts-module-publish-main", "cancel-in-progress: false",
                         "ref: ${{ github.sha }}", "fetch-depth: 0", "--current-main",
                         "GOWORK=off go test -race ./...", "GOWORK=off go build ./...",
                         "python3 -m unittest discover", "steps.changes.outputs.publish == 'true'"):
            self.assertIn(required, workflow)
        self.assertNotIn("pull_request_target:", workflow)
        self.assertNotIn("--force", workflow)

    def test_remote_tag_collision_is_never_overwritten(self):
        self.baseline()
        baseline = self.git("rev-parse", "HEAD")
        self.write("domain/input.go")
        self.save()
        self.remote()
        original_git = publication.git

        def competing_push(*args):
            if args[0] == "push":
                self.git("--git-dir", str(self.origin), "update-ref", "refs/tags/v0.1.112", baseline)
            return original_git(*args)

        with mock.patch.object(publication, "git", side_effect=competing_push):
            with self.assertRaises(subprocess.CalledProcessError):
                self.release()
        self.assertEqual(self.git("--git-dir", str(self.origin), "rev-parse", "v0.1.112"), baseline)

    def test_conflicting_fetched_tag_is_not_force_replaced(self):
        self.baseline()
        baseline = self.git("rev-parse", "v0.1.111")
        self.write("domain/input.go")
        target = self.save()
        self.remote()
        self.git("--git-dir", str(self.origin), "update-ref", "refs/tags/v0.1.111", target)
        with self.assertRaises(subprocess.CalledProcessError):
            self.release()
        self.assertEqual(self.git("rev-parse", "v0.1.111"), baseline)
        self.assertEqual(publication.latest_release(), "v0.1.111")


if __name__ == "__main__":
    unittest.main()
