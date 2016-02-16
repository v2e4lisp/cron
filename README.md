

# Cron

http://crontab.org/

## Usage

```golang
cron := NewCronJob()
hi := func() { fmt.Println("hello world") }
cron.Register("* * * * *", hi)
// every minute print "hello world"
```

