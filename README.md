# VJG Dienynas

https://vjgdiary.neglostyti.com/

This is a personal toy project: a modernized view on top of old school system.

It downloads data on your behalf:
* Your personal schedule from internal account with past lectures and assignments;
* Public schedule from https://vjg.edupage.org/timetable/

Then everything is merged and presented as single view, containing information about past lectures, which lesson is next, and homework tasks, sorted by priority. Homework for next day is highlighted separately.

## Notes to self

Lambda deployment is managed with SAM. When in doubt, delete CloudFormation stack and start over.

Cloud prerequisites: onboarding certificate from CloudFlare, and setting up SSL:strict rule for that specific domain in CF.

Maintenance tasks:

```
task lint
task deploy 
```