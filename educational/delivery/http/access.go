package http

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net"
	"strconv"
	"strings"
)

const (
	sessionUnlockedTasksKey = "unlocked_tasks"
	// maxUnlockedTasks keeps the session cookie small; the oldest unlocks are forgotten first.
	maxUnlockedTasks = 100
)

// unlockedTasks lists tasks the student entered passwords for, oldest first. The session cookie is signed
// but readable, so it keeps only task ids and password versions, never passwords or their hashes.
type unlockedTasks []unlockedTask

type unlockedTask struct {
	id      int
	version int
}

func (u unlockedTasks) IsUnlocked(subjectObjectId int, passwordVersion int) bool {
	for _, task := range u {
		if task.id == subjectObjectId {
			return task.version == passwordVersion
		}
	}
	return false
}

func readUnlockedTasks(c *gin.Context) unlockedTasks {
	value, _ := sessions.Default(c).Get(sessionUnlockedTasksKey).(string)
	var tasks unlockedTasks
	for _, item := range strings.Split(value, ",") {
		idPart, versionPart, ok := strings.Cut(item, ":")
		if !ok {
			continue
		}
		id, idErr := strconv.Atoi(idPart)
		version, versionErr := strconv.Atoi(versionPart)
		if idErr == nil && versionErr == nil {
			tasks = append(tasks, unlockedTask{id: id, version: version})
		}
	}
	return tasks
}

func saveUnlockedTask(c *gin.Context, subjectObjectId int, passwordVersion int) error {
	tasks := make(unlockedTasks, 0, maxUnlockedTasks)
	for _, task := range readUnlockedTasks(c) {
		if task.id != subjectObjectId {
			tasks = append(tasks, task)
		}
	}
	tasks = append(tasks, unlockedTask{id: subjectObjectId, version: passwordVersion})
	if len(tasks) > maxUnlockedTasks {
		tasks = tasks[len(tasks)-maxUnlockedTasks:]
	}

	items := make([]string, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, strconv.Itoa(task.id)+":"+strconv.Itoa(task.version))
	}

	session := sessions.Default(c)
	session.Set(sessionUnlockedTasksKey, strings.Join(items, ","))
	return session.Save()
}

// clientIP returns the address of the student's browser. In production the app sits behind Traefik,
// which appends the address it received the request from to X-Forwarded-For; earlier entries come
// from the client and can be forged, so only the last one is trusted, and only from a private network.
func clientIP(c *gin.Context) string {
	remote, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
	if err != nil {
		return c.Request.RemoteAddr
	}

	remoteIP := net.ParseIP(remote)
	forwarded := c.GetHeader("X-Forwarded-For")
	if remoteIP == nil || forwarded == "" || !isPrivate(remoteIP) {
		return remote
	}

	items := strings.Split(forwarded, ",")
	last := strings.TrimSpace(items[len(items)-1])
	if net.ParseIP(last) == nil {
		return remote
	}
	return last
}

func isPrivate(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate()
}
