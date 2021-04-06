package gitmodule

import (
	git "github.com/libgit2/git2go/v30"
)

type Blame struct {
	StartLineNo int     `json:"startLineNo"`
	EndLineNo   int     `json:"endLineNo"`
	Commit      *Commit `json:"commit"`
}

// BlameFile
func (repo *Repository) BlameFile(commit *Commit, filePath string) ([]*Blame, error) {
	rawRepo, err := repo.GetRawRepo()
	if err != nil {
		return nil, err
	}
	opts := git.BlameOptions{
		NewestCommit: commit.Git2Oid(),
		OldestCommit: nil,
		MinLine:      1,
		MaxLine:      10000,
	}
	blame, err := rawRepo.BlameFile(filePath, &opts)
	if err != nil {
		return nil, err
	}
	defer blame.Free()

	count := blame.HunkCount()

	blames := []*Blame{}
	if count > 0 {
		preHunk, err := blame.HunkByIndex(0)
		if err != nil {
			return nil, err
		}
		blameLineCommit, err := repo.GetCommit(preHunk.FinalCommitId.String())
		if err != nil {
			return nil, err
		}
		blames = append(blames, &Blame{
			StartLineNo: int(preHunk.FinalStartLineNumber),
			EndLineNo:   int(preHunk.FinalStartLineNumber),
			Commit:      blameLineCommit,
		})
		for i := 1; i < count; i++ {
			currentHunk, _ := blame.HunkByIndex(i)
			blameLineCommit, err := repo.GetCommit(currentHunk.FinalCommitId.String())
			if err != nil {
				return nil, err
			}
			blameCount := len(blames)
			//前后2个hunk的commit一样，自动合并
			if blameCount > 0 && blameLineCommit.ID == blames[blameCount-1].Commit.ID {
				blames[len(blames)-1].EndLineNo = int(currentHunk.FinalStartLineNumber) - 1
			} else {
				blames[len(blames)-1].EndLineNo = int(currentHunk.FinalStartLineNumber) - 1
				blames = append(blames, &Blame{
					StartLineNo: int(currentHunk.FinalStartLineNumber),
					EndLineNo:   int(currentHunk.FinalStartLineNumber),
					Commit:      blameLineCommit,
				})
			}

			preHunk = currentHunk
		}

	}

	return blames, nil
}
