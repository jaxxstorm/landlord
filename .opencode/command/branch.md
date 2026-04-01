---
description: Create and check out a new git branch
---

Create a new git branch and switch to it.

**Input**: The argument after `/branch` is the new branch name. If omitted, ask the user for the branch name.

**Steps**

1. **Get the branch name**

   If no branch name is provided, use the **AskUserQuestion tool** to ask for one.

   The branch name should be a short descriptive git branch name such as:
   - `feat/tailscale-auth`
   - `fix/api-timeout`
   - `chore/opencode-commands`

2. **Validate the branch name**

   Check that the name is valid for git and is not empty.

   If the name is clearly invalid, ask the user for a corrected name.

3. **Check for conflicts**

   Run these checks:
   - `git branch --list "<name>"`
   - `git branch --remotes "*/<name>"`

   If the branch already exists locally:
   - Tell the user
   - Ask whether to switch to the existing branch instead

   If the branch exists only on a remote:
   - Tell the user
   - Ask whether to create a local branch tracking it or choose another name

4. **Create and check out the branch**

   If the branch does not already exist, run:
   ```bash
   git checkout -b "<name>"
   ```

5. **Confirm the active branch**

   Run:
   ```bash
   git branch --show-current
   ```

   Report the active branch name back to the user.

**Output On Success**

```
## Branch Ready

**Current branch:** <name>
```

**Guardrails**
- Always ask for the branch name if it is not provided
- Do not overwrite or recreate an existing branch silently
- Prefer `git checkout -b` to create and switch in one step
- If a matching remote branch exists, ask before creating a different local branch state
