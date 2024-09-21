## drone-bark

### Usage

```
steps:
  - name: notification
    when:
      status:
        - success
    image: docker.test4x.com/xgfan/drone-bark:fe787e04
    settings:
      token:
        from_secret: bark_token
      title: "{DRONE_REPO} 发布成功"
      content: "{DRONE_COMMIT_MESSAGE}"
```

