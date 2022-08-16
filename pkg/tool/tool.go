package tool

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/jiuhuche120/jhttp"
)

const (
	WorkingDayApi = "https://timor.tech/api/holiday/info"
	GithubApi     = "https://api.github.com"
	TeambitionApi = "http://teambition.hyperchain.cn:8099"
)

var DayStatus bool = false

// IsWorkingDay the api is unstable, so we need to try again.
func IsWorkingDay() (bool, error) {
	var isWorkingDay bool
	if err := retry.Retry(func(attempt uint) error {
		today := time.Now().Format("2006-01-02")
		client := jhttp.NewClient(
			jhttp.AddHeader("Accept", "application/vnd.github.v3+json"),
		)
		workingDayUrl := WorkingDayApi + "/" + today
		resp, err := client.Get(workingDayUrl, nil)
		if err != nil {
			return err
		}
		if !resp.IsSuccess() {
			body, err := resp.Body()
			if err != nil {
				return err
			}
			return fmt.Errorf("error getting day status: %v", string(body))
		}
		var day Day
		err = resp.JsonUnmarshal(&day)
		if err != nil {
			return err
		}
		if day.Type.Type == 0 {
			isWorkingDay = true
			return nil
		}
		isWorkingDay = false
		return nil
	}, strategy.Wait(10*time.Second), strategy.Limit(5)); err != nil {
		return isWorkingDay, fmt.Errorf("get day status failed: %v", err)
	}
	return isWorkingDay, nil
}

func IsMerged(pull *PullRequest, token string) (bool, error) {
	client := jhttp.NewClient(
		jhttp.AddHeader("Accept", "application/vnd.github.v3+json"),
		jhttp.AddHeader("Authorization", "token "+token),
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

func GetAllPullRequests(token, repository, project string) ([]*PullRequest, error) {
	client := jhttp.NewClient(
		jhttp.AddHeader("Accept", "application/vnd.github.v3+json"),
		jhttp.AddHeader("Authorization", "token "+token),
	)
	prUrl := fmt.Sprintf("%s/repos/%s/%s/pulls", GithubApi, repository, project)
	resp, err := client.Get(prUrl, nil)
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
	var pulls []*PullRequest
	err = resp.JsonUnmarshal(&pulls)
	if err != nil {
		return nil, err
	}
	return pulls, nil
}

func GetTeambitionCookie(email string, password string) ([]*http.Cookie, error) {
	client := jhttp.NewClient()
	loginUrl := fmt.Sprintf("%v/login", TeambitionApi)
	resp, err := client.Get(loginUrl, nil)
	if err != nil {
		return nil, err
	}
	body, err := resp.Body()
	if err != nil {
		return nil, err
	}
	// get secret from html
	clientID, err := getClientID(string(body))
	if err != nil {
		return nil, err
	}
	tokenID, err := getTokenID(string(body))
	if err != nil {
		return nil, err
	}

	client = jhttp.NewClient(
		jhttp.AddHeader("Content-Type", "application/x-www-form-urlencoded"),
	)
	form := jhttp.NewXForm(
		jhttp.AddXFormParams("email", email),
		jhttp.AddXFormParams("password", password),
		jhttp.AddXFormParams("response_type", "session"),
		jhttp.AddXFormParams("token", tokenID),
		jhttp.AddXFormParams("client_id", clientID),
	)
	loginEmailUrl := fmt.Sprintf("%v/api/login/email", TeambitionApi)
	resp, err = client.Post(loginEmailUrl, form)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		body, err := resp.Body()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error getting teambtition cookie: %v", string(body))
	}
	return resp.Cookies(), nil
}

func GetTeambitionAllSubTask(cookie []*http.Cookie, project string, app string) ([]*SubTask, error) {
	var allSubTasks []*SubTask
	client := jhttp.NewClient()
	client.AddCookie(cookie)
	opts := []jhttp.ParamsOption{
		jhttp.AddParams("filter", url.QueryEscape("_commongroupId = null AND isDone = false AND taskType = story")),
	}
	taskUrl := fmt.Sprintf("%v/api/projects/%v/smartgroups/%v/tasks", TeambitionApi, project, app)
	resp, err := client.Get(taskUrl, nil, opts...)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	err = resp.JsonUnmarshal(&tasks)
	if err != nil {
		return nil, err
	}
	// teambition can't use not null filter
	opts = []jhttp.ParamsOption{
		jhttp.AddParams("filter", url.QueryEscape("_commongroupId = null AND isDone = false AND taskType = story AND _sprintId = null")),
	}
	resp, err = client.Get(taskUrl, nil, opts...)
	if err != nil {
		return nil, err
	}
	var nullTasks []Task
	err = resp.JsonUnmarshal(&nullTasks)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]bool)
	for _, task := range nullTasks {
		mp[task.ID] = true
	}
	var doingTasks []Task
	for _, task := range tasks {
		if !mp[task.ID] {
			doingTasks = append(doingTasks, task)
		}
	}
	for _, task := range doingTasks {
		opts := []jhttp.ParamsOption{
			jhttp.AddParams("_ancestorId", task.ID),
			jhttp.AddParams("withSubtasks", "true"),
		}
		subTaskUrl := fmt.Sprintf("%v/api/tasks", TeambitionApi)
		resp, err := client.Get(subTaskUrl, nil, opts...)
		if err != nil {
			return nil, err
		}
		var subTasks []*SubTask
		err = resp.JsonUnmarshal(&subTasks)
		if err != nil {
			return nil, err
		}
		for _, tk := range subTasks {
			tk.Url = fmt.Sprintf("%v/task/%v", TeambitionApi, tk.ID)
		}
		allSubTasks = append(allSubTasks, subTasks...)
	}
	// teambition's time location is UTC, so we need to convert it to CST
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return nil, err
	}
	for _, task := range allSubTasks {
		task.StartDate = task.StartDate.In(location)
		task.DueDate = task.DueDate.In(location)
	}
	return allSubTasks, nil
}

func getClientID(body string) (string, error) {
	str := regexp.MustCompile("data-clientid=\"(.*)\"").FindString(body)
	str = strings.ReplaceAll(str, "\"", "")
	split := strings.Split(str, "=")
	if len(split) != 2 {
		return "", fmt.Errorf("error get clientID")
	}
	return split[1], nil
}

func getTokenID(body string) (string, error) {
	str := regexp.MustCompile("data-clienttoken=\"(.*)\"").FindString(body)
	str = strings.ReplaceAll(str, "\"", "")
	split := strings.Split(str, "=")
	if len(split) != 2 {
		return "", fmt.Errorf("error get clientID")
	}
	return split[1], nil
}
