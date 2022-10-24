package tool

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jiuhuche120/jhttp"
	"github.com/jiuhuche120/jreminder/pkg/types"
	"github.com/tidwall/gjson"
)

const (
	WorkingDayApi = "https://timor.tech/api/holiday/info"
	GithubApi     = "https://api.github.com"
	TeambitionApi = "http://teambition.hyperchain.cn:8099"
)

// IsWorkingDay check day types by holiday data
func IsWorkingDay(holiday string) bool {
	today := time.Now().Format("01-02")
	result := gjson.Parse(holiday).Get("holiday." + today + ".holiday")
	return !result.Bool()
}

func IsMerged(pull *types.PullRequest, token string) (bool, error) {
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

func GetAllPullRequests(token, repository, project string) ([]*types.PullRequest, error) {
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
	var pulls []*types.PullRequest
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
	form := jhttp.NewXFormParams(
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

func GetTeambitionAllSubTask(cookie []*http.Cookie, project string, app string) ([]*types.SubTask, error) {
	var allSubTasks []*types.SubTask
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
	var tasks []types.Task
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
	var nullTasks []types.Task
	err = resp.JsonUnmarshal(&nullTasks)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]bool)
	for _, task := range nullTasks {
		mp[task.ID] = true
	}
	var doingTasks []types.Task
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
		var subTasks []*types.SubTask
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

func NewGithubMsg(pulls []*types.PullRequest, text string) types.Msg {
	var msg types.Msg
	msg.MsgType = "text"
	for i := 0; i < len(pulls); i++ {
		msg.Text.Content += "🔗:" + pulls[i].HtmlUrl + " " + text + " @" + pulls[i].DingTalk + "\n"
		msg.At.AtMobiles = append(msg.At.AtMobiles, pulls[i].DingTalk)
	}
	msg.At.IsAtAll = false
	return msg
}

func NewTeambitionMsg(tasks []*types.SubTask, text string) types.Msg {
	var msg types.Msg
	msg.MsgType = "text"
	for _, task := range tasks {
		msg.Text.Content += "🔗:" + task.Url + " " + text + " @" + task.DingTalk + "\n"
		msg.At.AtMobiles = append(msg.At.AtMobiles, task.DingTalk)
	}
	msg.At.IsAtAll = false
	return msg
}
