@@ -0,0 +1,10 @@
# Vault Editor

Bridge uses an encrypted vault to store persistent data. One of the parameters stored in this vault is the roll factor (between 0.0 and 1.0)

It can be built with `make vault-editor` in the bridge source code root directory.

Example usage:

Setting the rollout value:
```bash
$ ./bridge-rollout set -v=0.81
0.81
```
Note that the provided value will be clamped between 0 and 1.

```bash
$ ./bridge-rollout get
0.81
```
