package bapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFindPullRequests(t *testing.T) {

	Convey("Given authenticated client to bitbucket repository", t, func() {
		client := authenticatedClient()

		Convey("When status is 'PullRequestAll'", func() {
			allRequests := client.PullRequests(PullRequestAll)

			Convey("Number of pull requests should be 2", func() {
				So(len(allRequests.PullRequests), ShouldEqual, 2)
			})
		})

		Convey("When status is 'PullRequestOpen'", func() {
			openRequests := client.PullRequests(PullRequestOpen)

			Convey("Number of pull requests should be 2", func() {
				So(len(openRequests.PullRequests), ShouldEqual, 2)
			})
		})

		Convey("When status is 'PullRequestDeclined'", func() {
			declinedRequests := client.PullRequests(PullRequestDeclined)

			Convey("Number of pull requests should be 1", func() {
				So(len(declinedRequests.PullRequests), ShouldEqual, 1)
			})
		})
		
		Convey("When status is 'PullRequestMerged'", func() {
			declinedRequests := client.PullRequests(PullRequestMerged)

			Convey("Number of pull requests should be 1", func() {
				So(len(declinedRequests.PullRequests), ShouldEqual, 1)
			})
		})
	})
}

func TestFindPullRequest(t *testing.T) {

	Convey("Given authenticated client to bitbucket repository", t, func() {
		client := authenticatedClient()
		Convey("When quering for existing pull request", func() {
			pullRequest := client.PullRequest(1)
			
			Convey("Should return Pull request", func() {
				So(pullRequest.Id, ShouldEqual, 1)
				So(pullRequest.State, ShouldEqual, "OPEN")
				So(len(pullRequest.Participants), ShouldEqual, 2)
				So(pullRequest.Participants[0].Approved, ShouldBeFalse)
			})
		})
	})
}

func authenticatedClient() *BClient {
	client := BClient{
		ConsumerKey:    "MSxtPGXknnpg9BEkpG",
		ConsumerSecret: "wth45aqgcEVD3JgTxHCrJqucwUF9KXEL",
		Repo:           "funny",
		Username:       "tszpinda"}
	client.Authenticate()

	return &client
}
