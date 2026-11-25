"""Fail on unpinned workflow `uses:` references or untrusted `pull_request_target`.

External actions must name a full 40-character commit SHA. Only the trusted
review gate may run on `pull_request_target`; this is the fallback where GitHub
offers no Actions event policy for the repository. The check reads lines, not
parsed YAML, so it is repository hygiene rather than a security boundary.
"""

from __future__ import annotations

import argparse
import re
import sys
from collections.abc import Sequence
from dataclasses import dataclass
from pathlib import Path
from typing import Final

FULL_SHA_REF: Final = re.compile(r"^[0-9a-fA-F]{40}$")
USES_KEY: Final = re.compile(r"(?:^|[-{,])\s*(?:uses|'uses'|\"uses\")\s*:\s*(.+)$")
LOCAL_USES_PREFIXES: Final = ("./", "../")
NON_GITHUB_ACTION_PREFIXES: Final = ("docker://",)
TRUSTED_GATE_WORKFLOW: Final = "review-gate-trusted.yml"
GATED_EVENT: Final = "pull_request_target"
TOP_LEVEL_ON: Final = re.compile(r"^(?:on|'on'|\"on\")\s*:(.*)$")
GATED_EVENT_NAME: Final = re.compile(r"(?<![\w-])pull_request_target(?![\w-])")
COMMENT: Final = re.compile(r"(?:^|\s)#")


@dataclass(frozen=True)
class Violation:
    path: Path
    line: int
    value: str
    reason: str


def workflow_files(workflows_dir: Path) -> list[Path]:
    return sorted([*workflows_dir.glob("*.yml"), *workflows_dir.glob("*.yaml")])


def _uses_value(line: str) -> str | None:
    if line.lstrip().startswith("#"):
        return None
    match = USES_KEY.search(line)
    if match is None:
        return None
    value = match.group(1).split("#", 1)[0].strip().rstrip(",}").strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in "'\"":
        value = value[1:-1].strip()
    return value


def gated_event_lines(text: str) -> list[int]:
    """Lines of the top-level `on:` key, inline or nested, that name the gated event."""
    lines: list[int] = []
    in_on = False
    for number, line in enumerate(text.splitlines(), start=1):
        body = COMMENT.split(line, maxsplit=1)[0].rstrip()
        if not body.strip():
            continue
        if not body[0].isspace():
            match = TOP_LEVEL_ON.match(body)
            in_on = match is not None
            if match is None:
                continue
            body = match.group(1)
        elif not in_on:
            continue
        if GATED_EVENT_NAME.search(body):
            lines.append(number)
    return lines


def scan_workflows(workflows_dir: Path) -> list[Violation]:
    violations: list[Violation] = []
    for workflow_path in workflow_files(workflows_dir):
        text = workflow_path.read_text(encoding="utf-8")
        if workflow_path.name != TRUSTED_GATE_WORKFLOW:
            violations.extend(
                Violation(workflow_path, line, GATED_EVENT, f"only {TRUSTED_GATE_WORKFLOW} may run on it")
                for line in gated_event_lines(text)
            )
        for line_number, line in enumerate(text.splitlines(), start=1):
            value = _uses_value(line)
            if value is None or not value:
                continue
            if value.startswith(LOCAL_USES_PREFIXES + NON_GITHUB_ACTION_PREFIXES):
                continue
            if "@" not in value:
                violations.append(Violation(workflow_path, line_number, value, "missing @<commit-sha> ref"))
                continue
            if not FULL_SHA_REF.fullmatch(value.rsplit("@", 1)[1]):
                violations.append(
                    Violation(workflow_path, line_number, value, "ref is not a 40-character commit SHA")
                )
    return violations


def format_violations(violations: Sequence[Violation]) -> str:
    lines = ["GitHub Actions workflow policy violations found:"]
    lines.extend(f"- {item.path}:{item.line}: {item.value} ({item.reason})" for item in violations)
    lines.extend(
        (
            "",
            "Use a full 40-character commit SHA for external actions.",
            f"Only {TRUSTED_GATE_WORKFLOW} may be triggered by {GATED_EVENT}.",
        )
    )
    return "\n".join(lines)


def main(argv: Sequence[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("workflows_dir", nargs="?", default=".github/workflows", type=Path)
    args = parser.parse_args(argv)
    if not args.workflows_dir.is_dir() or not workflow_files(args.workflows_dir):
        print(f"workflow directory is missing or empty: {args.workflows_dir}", file=sys.stderr)
        return 2
    violations = scan_workflows(args.workflows_dir)
    if violations:
        print(format_violations(violations), file=sys.stderr)
        return 1
    count = len(workflow_files(args.workflows_dir))
    print(f"workflow action pinning and {GATED_EVENT} use OK: {count} file(s)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
