# lumber_bot
--
lumber_bot is a Telegram bot that uses the Reddit API to share visceral images
of Earth, Space, Cities, Architecture, and more with your group.

Add @kungfu_kenney_bot to your group to experience it today!


### Your Own Bot

This package allows you to setup your own bot running locally or on a remote
server to achieve the same goal.

1. First, register a simple, script Reddit app to get your client id and secret
(https://github.com/reddit/reddit/wiki/OAuth2-Quick-Start-Example#first-steps)

2. Next, create a Telegram bot (https://core.telegram.org/bots#6-botfather)

3. Fill in etc/default-serverconf.yaml with the required values

5. Make any changes to etc/default-subreddits.yaml that you would like

6. Finally, run the bot

    lumber_bot -serverconf etc/serverconf -subredditconf etc/subredditconf

The bot will serve a random, highly upvoted, image from the past week from the
list of supplied subreddits when someone uses the "/get" command
