"""Exercise the workflow policy with valid pins and realistic regressions."""

import os
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

from check_workflow_pinning import scan_workflows  # noqa: E402

SHA = "a" * 40


class WorkflowPolicyTest(unittest.TestCase):
    def scan(self, files: dict[str, str]) -> list[tuple[str, int]]:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name, text in files.items():
                (root / name).write_text(text, encoding="utf-8")
            return [(os.path.basename(v.path), v.line) for v in scan_workflows(root)]

    def test_rejects_tags_branches_and_missing_refs(self) -> None:
        workflow = (
            "steps:\n"
            "  - uses: actions/checkout@v4\n"
            "  - uses: actions/setup-python@main\n"
            "  - uses: owner/action\n"
            f"  - uses: 'owner/action@{SHA}' # release\n"
            "  - uses: ./local-action\n"
            "# - uses: ignored/comment@main\n"
        )
        self.assertEqual(self.scan({"ci.yml": workflow}), [("ci.yml", 2), ("ci.yml", 3), ("ci.yml", 4)])

    def test_pull_request_target_only_in_trusted_gate(self) -> None:
        files = {
            "review-gate-trusted.yml": "on:\n  pull_request_target:\n    types: [opened]\n",
            "block.yml": "name: x\non:\n  push:\n  pull_request_target:\njobs: {}\n",
            "inline.yml": "on: [push, pull_request_target]\n",
            "quoted.yml": "'on':\n  - 'pull_request_target'\n",
            "mention.yml": (
                "# pull_request_target is not used here\n"
                "on:\n"
                "  pull_request:  # not pull_request_target\n"
                "jobs:\n"
                "  a:\n"
                "    if: github.event_name != 'pull_request_target'\n"
            ),
        }
        self.assertEqual(
            sorted(self.scan(files)),
            [("block.yml", 4), ("inline.yml", 1), ("quoted.yml", 2)],
        )


if __name__ == "__main__":
    unittest.main()
