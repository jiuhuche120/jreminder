package tool

import (
	"fmt"
	"strings"
	"time"

	http "github.com/jiuhuche120/jhttp"
)

const (
	WorkingDayApi = "https://timor.tech/api/holiday/info"
	GithubApi     = "https://api.github.com"
)

func IsWorkingDay() (bool, error) {
	today := time.Now().Format("2006-01-02")
	client := http.NewClient(
		http.AddHeader("Accept", "application/vnd.github.v3+json"),
	)
	url := WorkingDayApi + "/" + today
	resp, err := client.Get(url, nil)
	if err != nil {
		return false, err
	}
	if !resp.IsSuccess() {
		body, err := resp.Body()
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("error getting day status: %v", string(body))
	}
	var day Day
	err = resp.JsonUnmarshal(&day)
	if err != nil {
		return false, err
	}
	if day.Type.Type == 0 {
		return true, nil
	}
	return false, nil
}

func IsMerged(pull PullRequest) (bool, error) {
	client := http.NewClient(
		http.AddHeader("Accept", "application/vnd.github.v3+json"),
	)
	resp, err := client.Get(pull.Url+"/merge", nil)
	if err != nil {
		return false, err
	}
	if !resp.IsSuccess() {
		body, err := resp.Body()
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("error getting pr status: %v", string(body))
	}
	body, err := resp.Body()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(body), "No Content"), nil
}

func GetAllPullRequests(token, repository, project string) ([]PullRequest, error) {
	client := http.NewClient(
		http.AddHeader("Accept", "application/vnd.github.v3+json"),
		http.AddHeader("Authorization", "token "+token),
	)
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", GithubApi, repository, project)
	resp, err := client.Get(url, nil)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		body, err := resp.Body()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error getting all pull requests: %v", string(body))
	}
	var pulls []PullRequest
	err = resp.JsonUnmarshal(&pulls)
	if err != nil {
		return nil, err
	}
	return pulls, nil
}
