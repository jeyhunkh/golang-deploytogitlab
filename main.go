package main

import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "os/exec"
        "regexp"

        "github.com/go-resty/resty/v2"

        tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const BOT_TOKEN string = "643726534658:ahgghfkdgjfkjhjgfhgf"  // Telegram bot Token here

var authUserList = []int64{1234567}  //User ID list here. Separated by comma

func main() {

        bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
        if err != nil {
                log.Panic(err)
        }

        bot.Debug = true

        log.Printf("Authorized on account %s", bot.Self.UserName)

        u := tgbotapi.NewUpdate(0)
        u.Timeout = 60

        updates := bot.GetUpdatesChan(u)

        for update := range updates {

                msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
                userId := update.Message.From.ID
                fmt.Println(userId)
                if !contains(authUserList, userId) {
                        msg.Text = "You are not Authorized"
                        bot.Send(msg)
                } else {

                        if update.Message == nil {
                                continue
                        }

                        if update.Message.IsCommand() {
                                //msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

                                switch update.Message.Command() {

                                case "status":
                                        msg.Text = "I'm ok."

                                case "help":
                                        msg.Text = "Send file to this channel.The file name must contain either dev/test."
                                default:
                                        msg.Text = "I don't know that command"
                                }

                                if _, err := bot.Send(msg); err != nil {
                                        log.Panic(err)
                                }

                        }

                        if update.Message.Document != nil { // If we got a Document message
                                log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

                                filename := update.Message.Document.FileName
                                fmt.Println("The file name is ", filename)
                                fileID := update.Message.Document.FileID
                                fmt.Println("The file ID is : ", fileID)
                                caption := update.Message.Caption
                                fmt.Println("The caption is :", caption)
                                var commitMessage string
                                if caption == "" {
                                        commitMessage = "Bot added files"
                                        fmt.Println("The commit message is :", commitMessage)

                                } else if caption != "" {
                                        commitMessage = caption
                                        fmt.Println("The commit message is :", commitMessage)
                                }

                                matchdev := regexp.MustCompile("(?i)dev")
                                if matchdev.MatchString(filename) {
                                        fmt.Println("File name contains 'dev' string.")
                                        branchName := "dev"
                                        saveBotFile(BOT_TOKEN, getFilePath(BOT_TOKEN, fileID), branchName)
                                        unzipFile(branchName)
                                        cloneRrepo(branchName)
                                        changeFiles(branchName)

                                        gitPushToRepo(branchName, commitMessage)
                                        removeFolder(branchName)
                                } else {
                                        fmt.Println("File name does not contain 'dev' string.")
                                }

                                matchtest := regexp.MustCompile("(?i)test")
                                if matchtest.MatchString(filename) {
                                        fmt.Println("File name contains 'test' string.")
                                        branchName := "test"
                                        saveBotFile(BOT_TOKEN, getFilePath(BOT_TOKEN, fileID), branchName)
                                        unzipFile(branchName)
                                        cloneRrepo(branchName)
                                        changeFiles(branchName)
                                        gitPushToRepo(branchName, commitMessage)
                                        removeFolder(branchName)
                                } else {
                                        fmt.Println("File name does not contain 'test' string.")
                                }

                        }
                }
        }
}

func getFilePath(bot_token string, fileID string) (realFilePath string) {
        type FilePathResponse struct {
                Ok     bool `json:"ok"`
                Result struct {
                        FileID       string `json:"file_id"`
                        FileUniqueID string `json:"file_unique_id"`
                        FileSize     int    `json:"file_size"`
                        FilePath     string `json:"file_path"`
                } `json:"result"`
        }

        url := "https://api.telegram.org/bot" + bot_token + "/getFile?file_id=" + fileID

        fmt.Println("The URL is :", url)
        resp, err := http.Get(url)
        if err != nil {
                fmt.Println("Couldnt get filepath", err)
        }

        body, readErr := ioutil.ReadAll(resp.Body)
        if readErr != nil {
                log.Fatal(readErr)
        }

        fmt.Println("The body is : ", string(body))

        var result FilePathResponse
        if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
                fmt.Println("Can not unmarshal JSON")
        }

        realFilePath = result.Result.FilePath
        fmt.Println("The real file path is: ", realFilePath)

        defer resp.Body.Close()
        return realFilePath
}

func saveBotFile(bot_token string, file_path string, branchname string) {

        url := "https://api.telegram.org/file/bot" + bot_token + "/" + file_path

        if branchname == "dev" {
                client := resty.New()
                _, err := client.R().
                        SetOutput("./dev/archive.zip").
                        Get(url)
                if err != nil {
                        panic(err)
                }
        } else if branchname == "test" {
                client := resty.New()
                _, err := client.R().
                        SetOutput("./test/archive.zip").
                        Get(url)
                if err != nil {
                        panic(err)
                }

        }

}

func unzipFile(branchname string) {
        //cmd := exec.Command("ls", "./")
        if branchname == "dev" {
                cmd := exec.Command("unzip", "./dev/archive.zip", "-d", "./dev")

                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err)
                }
                fmt.Println("Output: ", string(out))
        } else if branchname == "test" {
                cmd := exec.Command("unzip", "./test/archive.zip", "-d", "./test")

                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err)
                }
                fmt.Println("Output: ", string(out))

        }

}

func cloneRrepo(cloneBranch string) {

        if cloneBranch == "dev" {
                cmd := exec.Command("git", "clone", "-b", cloneBranch, "--single-branch", "https://user:gitlabtokenhere@gitlab/my-repo/my-project.git", "./dev/my-project")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Output: ", string(out))
        } else if cloneBranch == "test" {
                cmd := exec.Command("git", "clone", "-b", cloneBranch, "--single-branch", "https://user:gitlabtokenhere@gitlab/my-repo/my-project.git", "./test/my-project")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Output: ", string(out))
        }
}

func changeFiles(branchname string) {
        //exec.Command("shopt", "-s", "extglob")
        if branchname == "dev" {
                cmd := exec.Command("cp", "-R", "./dev/build/*", "./dev/my-project/")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Output: ", string(out))
        } else if branchname == "test" {
                cmd := exec.Command("cp", "-R", "./test/build/*", "./test/my-project/")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Output: ", string(out))

        }

}

func gitPushToRepo(pushBranch string, commitMessage string) {
        if pushBranch == "dev" {
                cmd := exec.Command("git", "-C", "./dev/my-project/", "add", "--all")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully addedd all files ", string(out))

                cmd = exec.Command("git", "-C", "./dev/my-project/", "commit", "-m", commitMessage) //"\"Bot added new files\""
                out, err = cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully committed ", string(out))

                cmd = exec.Command("git", "-C", "./dev/my-project/", "push", "origin", pushBranch)
                out, err = cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully pushed all files ", string(out))

        } else if pushBranch == "test" {
                cmd := exec.Command("git", "-C", "./test/my-project/", "add", "--all")
                out, err := cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully addedd all files ", string(out))

                cmd = exec.Command("git", "-C", "./test/my-project/", "commit", "-m", commitMessage)
                out, err = cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully committed ", string(out))

                cmd = exec.Command("git", "-C", "./test/my-project/", "push", "origin", pushBranch)
                out, err = cmd.Output()
                if err != nil {
                        fmt.Println("could not run command: ", err.Error())
                }
                fmt.Println("Successfully pushed all files ", string(out))
        }

}

func removeFolder(branchName string) {
        if branchName == "dev" {
                dir := "./dev"
                err := os.RemoveAll(dir)
                if err != nil {
                        fmt.Println(err)
                        return
                }
                fmt.Println("Directory dev removed successfully")
        } else if branchName == "test" {
                dir := "./test"
                err := os.RemoveAll(dir)
                if err != nil {
                        fmt.Println(err)
                        return
                }
                fmt.Println("Directory test removed successfully")
        }
}

func contains(s []int64, e int64) bool {
        for _, a := range s {
                if a == e {
                        return true
                }
        }
        return false
}
