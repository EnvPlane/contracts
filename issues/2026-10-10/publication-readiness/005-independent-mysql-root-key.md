# MySQL root initialization must not reuse the app credential

Status: contracts binding fixed locally; main/worker generator and runtime proof
remain publication prerequisites.

Security review found worker root initialization selecting the app PasswordKey.
MySQLRestoreItem now requires TargetRootPasswordKey, distinct from the app key;
plan digest and trusted MySQLRestoreTargetSecret.RootPasswordKey bind selection.

Implementation prompt: the trusted mysql-password-v1 generator creates independent
app and root values, with fixed root key MYSQL_ROOT_PASSWORD. Authenticate actual
Secret UID/managed labels/Secret plan digest before approving both key names.
Adapt RootPasswordKey into worker target initialization; never fall back to the
app key, overwrite generated credentials, grant source permissions or accept
arbitrary client/runtime key flags. Test missing/equal/substituted keys and values,
Secret replacement and prior-lifecycle evidence. Do not publish before main
accepts generator/executor integration, live proof and all required checks.
