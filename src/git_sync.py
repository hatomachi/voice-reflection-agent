import subprocess
import logging
from pathlib import Path

logger = logging.getLogger("voice_reflection.git_sync")

def run_git(cmd: list[str], cwd: Path) -> subprocess.CompletedProcess:
    return subprocess.run(
        cmd,
        cwd=str(cwd),
        capture_output=True,
        text=True,
        check=False
    )

def git_fetch(repo_path: Path) -> bool:
    """Outbound git fetch origin main"""
    res = run_git(["git", "fetch", "origin", "main"], repo_path)
    if res.returncode != 0:
        logger.warning(f"git fetch failed: {res.stderr.strip()}")
        return False
    return True

def has_remote_updates(repo_path: Path) -> bool:
    """Check if origin/main is ahead of HEAD"""
    res = run_git(["git", "rev-list", "HEAD..origin/main", "--count"], repo_path)
    if res.returncode == 0:
        count = int(res.stdout.strip() or "0")
        return count > 0
    return False

def git_pull_rebase(repo_path: Path) -> bool:
    """Pull remote changes with rebase"""
    res = run_git(["git", "pull", "--rebase", "origin", "main"], repo_path)
    if res.returncode != 0:
        logger.error(f"git pull --rebase failed: {res.stderr.strip()}")
        return False
    logger.info("Successfully pulled latest changes from origin/main")
    return True

def git_sync_inbound(repo_path: Path) -> bool:
    """Fetch and pull if remote has updates"""
    if not git_fetch(repo_path):
        return False
    if has_remote_updates(repo_path):
        logger.info("Remote changes detected on origin/main. Pulling...")
        return git_pull_rebase(repo_path)
    return True

def git_commit_and_push(repo_path: Path, paths: list[str], commit_message: str) -> bool:
    """Add, commit, and push specified directories or files to origin main"""
    # 1. Add all changes (including deletions) in specified paths
    add_cmd = ["git", "add", "-A"] + paths
    res = run_git(add_cmd, repo_path)
    if res.returncode != 0:
        logger.error(f"git add failed: {res.stderr.strip()}")
        return False

    # 2. Check if there are staged changes
    res = run_git(["git", "diff", "--staged", "--name-only"], repo_path)
    if not res.stdout.strip():
        logger.info("No staged changes to commit.")
        return True

    # 3. Commit
    res = run_git(["git", "commit", "-m", commit_message], repo_path)
    if res.returncode != 0:
        logger.error(f"git commit failed: {res.stderr.strip()}")
        return False
    logger.info(f"Committed: {commit_message}")

    # 4. Push
    res = run_git(["git", "push", "origin", "main"], repo_path)
    if res.returncode != 0:
        logger.error(f"git push failed: {res.stderr.strip()}")
        return False
    logger.info("Successfully pushed to origin/main")
    return True
