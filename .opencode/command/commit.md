---
description: Analyze changes, suggest a conventional commit, and confirm before committing
---

Create a git commit from the current worktree.

**Input**: Optionally specify a commit message after `/commit`. If omitted, inspect the current changes and suggest a conventional commit message.

**Steps**

1. **Inspect repository state**

   Run these git commands in parallel:
   - `git status --short`
   - `git diff --staged`
   - `git diff`
   - `git log -5 --oneline`

   Use the results to understand:
   - Which files are modified, staged, or untracked
   - The overall purpose of the change
   - The recent commit style used in this repository

2. **Handle empty state**

   If there are no staged, unstaged, or untracked changes:
   - Stop
   - Tell the user there is nothing to commit

3. **Determine what should be committed**

   Review the changed files and summarize the change in plain language.

   If the user provided a commit message after `/commit`, treat it as the preferred message but still validate that it matches the actual changes.

   If the message is missing or too vague, draft a conventional commit message that reflects the change accurately.

   Prefer commit types such as:
   - `feat`: new functionality
   - `fix`: bug fix
   - `refactor`: internal restructuring without behavior change
   - `test`: test-only changes
   - `docs`: documentation-only changes
   - `chore`: maintenance or tooling updates

   Keep the subject concise and focused on why the change exists.

4. **Protect against unsafe commits**

   Do not commit likely secret files such as `.env`, `*.pem`, `*.key`, or credential dumps.

   If such files appear in the change set:
   - Warn the user clearly
   - Ask whether to proceed without those files
   - Do not stage them by default

5. **Prompt for confirmation**

   Show:
   - A short summary of the changes to be committed
   - The proposed conventional commit message
   - Which files will be included

   Then use the **AskUserQuestion tool** to ask for confirmation with these options:
   - `Commit now (recommended)`
   - `Edit message`
   - `Cancel`

   If the user chooses `Edit message`, ask for the revised message and use that for the commit.

6. **Stage and commit**

   If the user confirms, stage the relevant modified and untracked files, excluding anything identified as sensitive.

   Then run:
   ```bash
   git add <relevant files>
   git commit -m "<message>"
   ```

   After the commit, run `git status --short` to verify the result.

7. **Report outcome**

   On success, show:
   - The final commit message
   - The created commit hash
   - Remaining uncommitted files, if any

**Output On Success**

```
## Commit Complete

**Message:** <type(scope): summary>
**Commit:** <hash>

Committed files:
- <file>
- <file>
```

**Output On Cancel**

```
## Commit Cancelled

No commit was created.
```

**Guardrails**
- Always inspect the worktree before proposing a message
- Always prompt before creating the commit
- Suggest a conventional commit message when one is not provided
- Do not commit likely secret files by default
- Do not create an empty commit
- Do not push after committing unless the user explicitly asks
