# Reddit comment checker
This chrome extension checks the top level of all comments in a reddit post for potentially similar comments. Similar comments are displayed and you can upvote them instead of posting duplicate content!

Simply type out your comment and press the "Check for duplicates" button.

## Lambda function
The meat and potatoes of this extension are handled by an aws lambda written in Go.

To deploy your function you'll need to set up an "agentfile" in the comment_analyzer folder. A template has been provided. [Refer here to get your creds for that.](https://github.com/reddit-archive/reddit/wiki/oauth2)

You can deploy the function however you like, I personally use [Apex](http://apex.run/) and have included an example function.json, just supply the role you'd like your function to take.

## Chrome extension
The extension is pretty lightweight, it just calls the lambda function and displays the results.

You can build the extension by running npm install && npm run build

You'll notice that we're building something that's pretty small. Why do that, you ask? Because I originally thought I would handle more stuff in the extension, I answer. I'll probably rework all this in the future to make it less goofy.

### TODO:
- Add support for new reddit redesign
- Improve similarity recognition. It's _ok_ right now but not _great_
- Firefox support? At a quick glance it looks like everything should be identical for firefox. However I booted it up & it didn't work.
