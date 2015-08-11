package bapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPullRequestStatus(t *testing.T) {

	Convey("Given pull request status", t, func() {
		pullRequest := PullRequestOpen
		
		Convey("When String() method is called", func() {
			pullRequestString := pullRequest.String()
			
			Convey("value should be 'OPEN'", func() {
				So(pullRequestString, ShouldEqual, "OPEN")
			})

		})
	})
}
