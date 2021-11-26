package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v28/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_github "github.com/mendersoftware/integration-test-runner/client/github/mocks"
)

func TestBotHasAlreadyCommentedOnPR(t *testing.T) {
	type returnValues struct {
		issueComments []*github.IssueComment
		error         error
	}
	commentString := github.String(", Let me know if you want to start the integration pipeline by mentioning me and the command \"")
	conf := &config{
		githubOrganization: "mendersoftware",
	}
	testCases := map[string]struct {
		pr             *github.PullRequestEvent
		expectedResult bool
		returnVals     returnValues
	}{
		"Bot has not commented": {
			pr: &github.PullRequestEvent{
				PullRequest: &github.PullRequest{
					Merged: github.Bool(false),
				},
				Repo: &github.Repository{
					Name: github.String("I am not the bot"),
					Owner: &github.User{
						Name: github.String("mendersoftware"),
					},
				},
				Number: github.Int(6),
			},
			returnVals: returnValues{
				issueComments: nil,
				error:         errors.New("Failed to retrieve the comments"),
			},
			expectedResult: false,
		},
		"Bot has commented": {
			pr: &github.PullRequestEvent{
				PullRequest: &github.PullRequest{
					Merged: github.Bool(false),
				},
				Repo: &github.Repository{
					Name: commentString,
					Owner: &github.User{
						Name: github.String("mendersoftware"),
					},
				},
				Number: github.Int(6),
			},
			returnVals: returnValues{
				issueComments: []*github.IssueComment{
					{
						Body: commentString,
					},
				},
				error: nil,
			},
			expectedResult: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mclient := &mock_github.Client{}
			defer mclient.AssertExpectations(t)

			mclient.On("ListComments",
				mock.MatchedBy(func(ctx context.Context) bool {
					return true
				}),
				*tc.pr.Repo.Owner.Name,
				*tc.pr.Repo.Name,
				*tc.pr.Number,
				mock.MatchedBy(func(*github.IssueListCommentsOptions) bool {
					return true
				}),
			).Return(tc.returnVals.issueComments, tc.returnVals.error)

			log := logrus.NewEntry(logrus.StandardLogger())
			assert.Equal(t, tc.expectedResult, botHasAlreadyCommentedOnPR(log, mclient, tc.pr, *commentString, conf))
		})
	}
}

func TestStartPipeline(t *testing.T) {
	testCases := map[string]struct {
		branchName     string
		expectedResult bool
	}{
		"start pipeline 1": {
			branchName:     "master",
			expectedResult: true,
		},
		"start pipeline 2": {
			branchName:     "staging",
			expectedResult: true,
		},
		"start pipeline 3": {
			branchName:     "production",
			expectedResult: true,
		},
		"start pipeline 4": {
			branchName:     "3.1.",
			expectedResult: true,
		},
		"start pipeline 5": {
			branchName:     "3.1.x",
			expectedResult: true,
		},
		"start pipeline 6": {
			branchName:     "pr_1",
			expectedResult: true,
		},
		"do not start pipeline 1": {
			branchName:     "other-branch",
			expectedResult: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedResult, shouldStartPipeline(tc.branchName))
		})
	}
}
