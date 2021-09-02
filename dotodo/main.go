package main

import (
  "flag"
  "fmt"
  "log"
  "os"
  "path"
  "runtime"
  "strings"
  "time"

  "gorm.io/gorm"
  "gorm.io/gorm/sqlite"
)

var (
  db *gorm.DB
  fileLogger *log.Logger
  logger = log.New(os.Stderr, "", 0)
)

func init() {
  _, thisFile, _, ok := runtime.Caller(0)
  if !ok {
    logger.Fatal("error loading source file")
  }
  thisDir := path.Dir(thisFile)

  logFile, err := os.OpenFile(
    path.Join(thisDir, "dotodo.log"),
    os.O_CREATE|os.O_APPEND|os.O_WRONLY,
    0644,
  )
  if err != nil {
    logger.Fatalf("error opening log file: %v", err)
  }
  fileLogger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)

  db, err = gorm.Open(path.Join(thisDir, sqlite.Open("dotodo.db")), &gorm.Config{})
  if err != nil {
    logBoth("error opening database: %v", err)
  }
}

func main() {
  taskName := flag.String("name", "", "Task name")
  taskDesc := flag.String("desc", "", "Task description")
  createTask := flag.Bool("create", false, "Create a new task")
  printTask := flag.String("print-task", "", "Print task with given name")
  printTasks := flag.Bool("print", false, "Print uncompleted tasks")
  printCompletedTasks := flag.Bool("print-fin", false, "Print completed tasks")
  printAllTasks := flag.Bool("print-all", false, "Print all tasks (completed and not)")
  completeTask := flag.String("fin", "", "Mark task as completed")
  uncompleteTask := flag.String("unfin", "", "Mark task as uncompleted")
  updateTask := flag.Bool("update", false, "Update a task")
  replaceTask := flag.Bool("replace", false, "Replace a task")
  deleteTask := flag.String("delete", "", "Delete a task")
  deleteCompletedTasks := flag.Bool("delete-fin", false, "Delete all completed tasks")
  deleteAllTasks := flag.Bool("delete-all", false, "Delete all tasks")
  flag.Parse()

  flagsPassed := make(map[string]bool, 10)
  flag.Visit(func(f *flag.Flag)) {
    flagsPassed[f.Name] = true
  })

  name, desc := *taskName, *taskDesc
  switch {
    case *createTask:
      if name == "" || desc == "" {
        logger.Fatal("must provide task name and description")
      }
      CreateTask(name, desc)
    case flagsPassed["print-task"]:
      if *printTask != "" {
        name = *printTask
      }
      if name == "" {
        logger.Fatal("must provide task name")
      }
      PrintTask(name)
    case *printTasks:
      PrintTasks()
    case *printCompletedTasks:
      PrintCompletedTasks()
    case *printAllTasks:
      PrintAllTasks()
    case flagsPassed["fin"]:
      if *completeTask != "" {
        name = *completedTask
      }
      if name == "" {
        logger.Fatal("must provide task name")
      }
      CompleteTask(name)
    case flagsPassed["unfin"]:
      if *uncompleteTask != "" {
        name = *uncompleteTask
      }
      if name == "" {
        loger.Fatal("must provide task name")
      }
      UncompleteTask(name)
    case *updateTask:
      if name == "" || desc == "" {
        logger.Fatal("must provide task name and description")
      }
      UpdateTask(name, desc)
    case *replaceTask:
      if name == "" || desc == "" {
        logger.Fatal("must provide task name and description")
      }
      ReplaceTask(name, desc)
    case flagsPassed["delete"]:
      if *deleteTask != nil {
        nmame = *deleteTask
      }
      if name == "" {
        logger.Fatal("must provide task name")
      }
      DeleteTask(name)
    case *deleteCompletedTasks:
      DeleteCompletedTasks()
    case *deleteAllTasks:
      DeleteAllTasks()
  }
}

type Task struct {
  Name string `gorm:"unique"`
  Description string
  CompletedAt int64
}

func NewTask(name, desc string) *Task {
  return &Task{
    Name: name,
    Description: desc,
  }
}

func CreateTask(name, desc string) {
  res := db.Create(NewTask(name, desc))
  if res.Error != nil {
    if strings.Contains(res.Error.Error(), "unique") {
      logBoth("task already exists")
    }
    logBoth("error creating task: %v", res.Error)
  }
}

func PrintTask(name string) {
  var task Task
  res := db.Where("name = ?", name).First(&task)
  if res.Error != nil {
    logBoth("error retrieving task: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Printf("no task named \"%s\"", name)
    return
  }
  msg := task.Name
  if task.CompletedAt == 0 {
    msg += " (uncompleted)"
  } else {
    msg += fmt.Sprintf(" (completed at: %d)", task.CompletedAt)
  }
  fmt.Printf("%s\n\t%s\n", msg, task.Description)
}

func PrintTasks() {
  var tasks []Task
  res := db.Where("completed_at = ?", 0).Find(&tasks)
  if res.Error != nil {
    logBoth("error retrieving tasks (uncompleted): %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("no tasks (uncompleted)")
    return
  }
  for _, task := range tasks {
    fmt.Printf("%s (uncompleted)\n\t%s\n", task.Name, task.Description)
  }
}

func PrintCompletedTasks() {
  var tasks []Task
  res := db.Where("completed_at > ?", 0).Find(&tasks)
  if res.Error != nil {
    logBoth("error retrieving completed tasks: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("no completed tasks")
    return
  }
  for _, task := range tasks {
    fmt.Printf("%s (completed at: %d)\n\t%s\n", task.Name, task.CompletedAt, task.Description)
  }
}

func PrintAllTasks() {
  var tasks []Task
  res := db.Find(&tasks)
  if res.Error != nil {
    logBoth("error retrieving all tasks: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("no tasks")
    return
  }
  for _, task := range tasks {
    msg := task.Name
    if task.CompletedAt == 0 {
      msg += " (uncompleted)"
    } else {
      msg += fmt.Sprintf(" (completed at: %d", task.CompletedAt)
    }
    fmt.Printf("%s\n\t%s\n", msg, task.Description)
  }
}

func CompleteTask(name string) {
  res := db.Model(&Task{}).Where("name = ?", name).Where("completed_at = ?", 0).Update("completed_at", time.Now().Unix())
  if res.Error != nil {
    logBoth("error completing task: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("task doesn't exist")
  }
}

func UncompleteTask(name string) {
  res := db.Model(&Task{}).Where("name = ?", name).Update("completed_at", 0)
  if res.Error != nil {
    logBoth("error uncompleting task: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("task doesn't exist")
  }
}

func UpdateTask(name, desc string) {
  res := db.Model(&Task{}).Where("name = ?", name).Update("description", desc)
  if res.Error != nil {
    logBoth("error updating task: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("task doesn't exist")
  }
}

func ReplaceTask(name, desc string) {
  task := &Task{Name: name}
  fields := map[string]interface{}{"name": name, "description", desc, "completed_at", 0}
  res := db.Model(task).Updates(fields)
  if res.Error != nil {
    logBoth("error replacing row: %v", res.Error)
  } else if res.RowsAffected == 0 {
    res := db.Create(NewTask(name, desc))
    if res.Error != nil {
      logBoth("error replacing (creating) row: %v", res.Error)
    }
  }
}

func DeleteTask(name string) {
  res := db.Where("name = ?", name).Delete(&Task)
  if res.Error != nil {
    logBoth("error deleting task: %v", res.Error)
  } else if res.RowsAffected == 0 {
    logger.Println("task doesn't exist")
  }
}

func DeleteCompletedTasks() {
  res := db.Where("completed_at > ?", 0).Delete(&Task{})
  if res.Error != nil {
    logBoth("error deleting completed tasks: %v", res.Error)
  }
}

func DeleteAllTasks() {
  res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Task)
  if err != nil {
    logBoth("error deleting all tasks: %v", res.Error)
  }
}

func logBoth(format string, items ...interface{}) {
  msg := fmt.Sprintf(format, items...)
  logger.Print(msg)
  fileLogger.Fatal(msg)
}
