ALTER TABLE dice_issues ADD `reopen_count` int(11) NOT NULL DEFAULT 0 COMMENT 'issue reopen count';

UPDATE dice_issues LEFT JOIN dice_issue_state ON dice_issues.state = dice_issue_state.id set reopen_count = 1 where dice_issue_state.belong = 'REOPEN';