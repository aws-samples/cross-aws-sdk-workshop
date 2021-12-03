# Lambda Design

## Add RSS API Gateway
- Invoke AddPodcast Lambda Handler
- return Episodes that were found

-------------------------------------

Example Importing single episode directly

```sh
curl -X POST "$API_URL/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_episode": {"title": "I have an idea", "description": "But what could it be?", "url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3"}}'

curl -X POST "$API_URL/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_episode": {"title": "I have another idea", "description": "But what could that one be?", "url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3", "content_type": "audio/mpeg"}}'
```

Example importing latest 2 episodes from RSS feed

```sh
curl -X POST "$API_URL/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_rss_feed": {"title": "AWS Podcast", "url": "https://d3gih7jbfe3jlq.cloudfront.net/aws-podcast.rss", "max_num_episodes": 2}}'
```

Example of importing episode with now media URL

```sh
curl -X POST "$API_URLom/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_episode": {"title": "Episode with no media", "description": "But what could it be?"}}'
```

------------------------------------


## Process episode state machine:
- Input:
    - episode, JSON object
        - ID (DDB record id)
        - Media URL of episode
        - Media content type

Example episode state machine input

```json
{
    "episode": {
      "id": "1234-5678-980",
      "name": "Why AWS for Genomics",
      "media_url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3",
      "media_content_type": "audio/mpeg",
      "status": "pending"
    }
}
```

Optionally add `transcribe_job_id` to `episode` to reuse existing
transcription job output.
`"transcribe_job_id": "503b2007-49fd-41fa-a410-3bbb9a51bac2"`

### Update Episode status
- Update Status of episode in DDB to downloaded, pending episode S3 upload

### Upload Podcast
- Download podcast episode, and upload to S3

### Update Episode status
- Update Status of episode in DDB to downloaded, pending transcription

### Start Transcription
- Start transcription job

### Check Transcription
- Check status of transcription job

### Process transcription
- Add transcription output text to DDB item 

### Complete

### Failure
### Update Episode status
- Update Status of episode in DDB to failed


-------------------------------------
### Setup API URL variable:

```
eval `make export-api-url`
```

### List Podcasts:

```
curl -i -X GET "${API_URL}/podcast"
```

### Get Podcast:

```
curl -i -X GET "${API_URL}/podcast/{id}"
```

### Play Podcast:

```
curl -i -X GET "${API_URL}/podcast/{id}/play?content=text|media"
```

### Import Podcast RSS Feed:

```
curl -i -X POST "${API_URL}/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_rss_feed": {"url": "https://d3gih7jbfe3jlq.cloudfront.net/aws-podcast.rss", "max_num_episodes": 2}}'
```

Sample Feeds:
* https://d3gih7jbfe3jlq.cloudfront.net/aws-podcast.rss

### Import Podcast Episode:

```
curl -i -X POST "${API_URL}/podcast" \
	-H "Content-Type: applcation/json" \
	-d '{"import_episode": {"title": "I have another idea", "description": "But what could that one be?", "url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3", "content_type": "audio/mpeg"}}'
```

Episode URLs:
* https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3
* https://rss.art19.com/episodes/299a935e-314b-4e68-a2d7-90561d60abfd.mp3

### Combined command for waiters example:

```
curl -i -X POST "${API_URL}/podcast" \
	-H "Content-Type: application/json" \
	-d '{"import_episode": {"id": "play-test-2", "title": "AWS for Genomics", "url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3", "content_type": "audio/mpeg"}}' && \
curl -i -X GET "${API_URL}/podcast/play-test-2/play?content=media"
```

--- 
```
http --raw '{"import_episode": {"id": "waiter-test-2", "title": "AWS for Genomics", "url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3", "content_type": "audio/mpeg"}}' \
    POST "${API_URL}/podcast" && \
http --follow "${API_URL}/podcast/waiter-test-2/play?content=media"
```

