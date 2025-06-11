# Current Task: Fansly Tag Statistics Tracker

## Completed Tasks ✓

1. **Database Schema** - Set up SQLite database with tags, tag_history, and tag_requests tables
2. **Backend API** - Created RESTful API endpoints for tag management
3. **Fansly API Client** - Implemented client to fetch tag data from Fansly
4. **Frontend UI** - Built SvelteKit app with:
   - Sortable, filterable, paginated table
   - Tag history visualization with Chart.js
   - Date range picker for historical data
   - Tag request dialog for users

## Next Steps: Worker System for Tag Discovery & Tracking

### 1. Create Worker Infrastructure ✓
- [x] Set up a background job system using Bun's built-in timer functions
- [x] Create a worker manager that can:
  - Schedule periodic jobs
  - Handle job failures and retries
  - Log job execution

### 2. Tag Discovery Worker
- [ ] Create `src/workers/tag-discovery.ts` that:
  - Fetches recent posts from multiple tags
  - Extracts new tags from post content using regex
  - Checks if tags already exist in database
  - Adds new tags to the database
  - Respects Fansly API rate limits

### 3. Tag Update Worker ✓
- [x] Create `src/workers/tag-updater.ts` that:
  - Fetches all tracked tags from database
  - Updates view counts for each tag
  - Creates history records for changes
  - Runs daily (configurable)
  - Handles API errors gracefully

### 4. Worker Configuration ✓
- [x] Add environment variables:
  - `WORKER_UPDATE_INTERVAL` (default: 24 hours)
  - `WORKER_DISCOVERY_INTERVAL` (default: 6 hours)
  - `FANSLY_API_RATE_LIMIT` (requests per minute)
  - `WORKER_ENABLED` (to disable in development)

### 5. Worker Monitoring ✓
- [x] Add worker status endpoint: `GET /api/workers/status`
- [x] Track last run time, next run time, and status
- [x] Add manual trigger endpoints (with auth):
  - `POST /api/workers/discovery/trigger` (TODO: auth)
  - `POST /api/workers/update/trigger` (TODO: auth)

### 6. Database Optimizations ✓ (Partial)
- [x] Add indexes for frequently queried fields
- [x] Add a `last_checked` timestamp to tags table
- [ ] Consider implementing soft deletes for tags

### 7. Frontend Enhancements
- [ ] Add worker status indicator in the UI
- [ ] Show "last updated" timestamp for each tag
- [ ] Add loading states while data is being fetched
- [ ] Implement real-time updates using SSE or WebSockets

### 8. Error Handling & Logging
- [ ] Implement structured logging for workers
- [ ] Add error tracking for failed API requests
- [ ] Create alerts for critical failures
- [ ] Store error logs in database

### 9. Performance Optimizations
- [ ] Implement caching for frequently accessed data
- [ ] Batch database operations
- [ ] Add request queuing to respect rate limits
- [ ] Consider using a job queue (Bull/BullMQ alternative for Bun)

### 10. Additional Features
- [ ] Tag categories/grouping
- [ ] Export data to CSV/JSON
- [ ] Tag popularity trends
- [ ] Email notifications for tracked tags
- [ ] API authentication for public endpoints

## Implementation Priority

1. **High Priority** (Do First):
   - Basic worker infrastructure
   - Tag update worker (to keep existing tags fresh)
   - Worker status endpoint

2. **Medium Priority**:
   - Tag discovery worker
   - Worker monitoring UI
   - Error handling improvements

3. **Low Priority**:
   - Advanced features
   - Performance optimizations
   - Export functionality

## Technical Considerations

- Use Bun's native capabilities where possible
- Keep workers lightweight and stateless
- Implement proper error boundaries
- Consider memory usage for long-running processes
- Plan for horizontal scaling if needed

## Testing Strategy

- Unit tests for worker logic
- Integration tests for API interactions
- Mock Fansly API responses for testing
- Test rate limiting behavior
- Test error recovery mechanisms

## Fansly Requests

Get view count for a tag:

Request:

`GET https://apiv3.fansly.com/api/v1/contentdiscovery/media/tag?tag=young&ngsw-bypass=true`

Response:

```json
{
  "success": true,
  "response": {
    "mediaOfferSuggestionTag": {
      "id": "436274778790174756",
      "tag": "young",
      "description": "",
      "viewCount": 73399232,
      "flags": 0,
      "createdAt": 1665510372000
    },
    "aggregationData": {}
  }
}
```

Get posts for a tag:

Request:

`GET https://apiv3.fansly.com/api/v1/contentdiscovery/media/suggestionsnew?before=0&after=0&tagIds=436274778790174756&limit=25&offset=0&ngsw-bypass=true`

Response:

```json
{
  "success": true,
  "response": {
    "mediaOfferSuggestions": [],
    "aggregationData": {
      "accounts": [],
      "accountMedia": [],
      "accountMediaBundles": [],
      "posts": [
        {
          "id": "757666999579979776",
          "accountId": "311020182975819776",
          "content": "I love pink 💗\n\n#18 #teen #young #anal #asshole #bigass #naturalbody #brunette #blowjob #cameltoe #slim #pussyfuck #fit #gym #suck #wetting #tattoed #wetpussy #fetish #latin #barelylegal",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1742136249,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "757666999579979776",
              "pos": 0,
              "contentType": 1,
              "contentId": "756942637801418752"
            }
          ],
          "likeCount": 31,
          "replyCount": 9,
          "wallIds": [],
          "mediaLikeCount": 793,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "761542283433615360",
          "accountId": "600002651043667968",
          "content": "Stepsister is waiting for me to feed her with my dick😈🤭\n\n\n@Mollyholy \n#stepsister #cute #amateur #teen #barelylegal #young #18 #skinny #bj #blowjob",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743060188,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "761542283433615360",
              "pos": 0,
              "contentType": 1,
              "contentId": "761542280120119296"
            }
          ],
          "likeCount": 48,
          "replyCount": 2,
          "wallIds": [],
          "mediaLikeCount": 932,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": [
            {
              "start": 60,
              "end": 69,
              "handle": "mollyholy",
              "accountId": "588052097270816768"
            }
          ]
        },
        {
          "id": "761793401065578496",
          "accountId": "681515416559820801",
          "content": "little stepsister having an afternoon nap .. you know what will happen next...\n\n#18 #petite #cute #stepsister #teen #teensex #pussy #smalltits #young #schoolgirl #nude #barelylegal ",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743120059,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "761793401065578496",
              "pos": 0,
              "contentType": 1,
              "contentId": "761753120865792000"
            }
          ],
          "likeCount": 2,
          "wallIds": [],
          "mediaLikeCount": 207,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "761959483848011776",
          "accountId": "681515416559820801",
          "content": "little stepsister taught she's home alone .. so I surprised her while she's changing clothes after school...\n\n#18 #petite #cute #stepsister #teen #teensex #skinny #smalltits #young #schoolgirl #barelylegal #babygirl ",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743159657,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "761959483848011776",
              "pos": 0,
              "contentType": 1,
              "contentId": "761927720320901120"
            }
          ],
          "likeCount": 2,
          "wallIds": [],
          "mediaLikeCount": 316,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "762013798256091136",
          "accountId": "284843717079085056",
          "content": "would you join if you saw us doing this 🤭💗\n\n@alexisaevans \n\n#fyp #young #cute #busty #petite #ass  #pawg #feet #bigboobs #wet #ass #tits #boobs #natural #nipples #flexible #naked #asshole #sexy #sexygirl #fuck #fucking #holes #girls #sexygirls #lesbian #car #outside",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743172606,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "762013798256091136",
              "pos": 0,
              "contentType": 1,
              "contentId": "762013797182353409"
            }
          ],
          "likeCount": 106,
          "replyCount": 11,
          "wallIds": [],
          "mediaLikeCount": 855,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": [
            {
              "start": 46,
              "end": 58,
              "handle": "alexisaevans",
              "accountId": "285234580909203456"
            }
          ]
        },
        {
          "id": "762352106953777153",
          "accountId": "636133253325004801",
          "content": "RAWR! I'm a naughty lil fucktoy 🤭 I play with fur toy while stepDaddy stretches my tiny pussy with his big hard cock and dumps his cum inside my pink wet hole.. 💦🥹\n\nFull video here: https://fansly.com/post/715694610512355329\n\n#little #roleplay #innocent #creampie #cumslut #petite #stepdaughter #young #taboo #freeuse",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743253265,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "762352106953777153",
              "pos": 0,
              "contentType": 1,
              "contentId": "762273607169617920"
            },
            {
              "postId": "762352106953777153",
              "pos": 1,
              "contentType": 7100,
              "contentId": "731636396627865601"
            },
            {
              "postId": "762352106953777153",
              "pos": 2,
              "contentType": 42001,
              "contentId": "762273600118992896"
            }
          ],
          "likeCount": 3,
          "replyCount": 1,
          "wallIds": [],
          "mediaLikeCount": 136,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "762500565509877760",
          "accountId": "469568019794767872",
          "content": "which 🕳️ do you choose? \n\n#tiktok #pussy #pussylips #asshole #spreading #young #cute #brunette #skinny",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743288660,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "762500565509877760",
              "pos": 0,
              "contentType": 1,
              "contentId": "762500564977197056"
            }
          ],
          "likeCount": 42,
          "replyCount": 2,
          "wallIds": [],
          "mediaLikeCount": 517,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "762714508241154048",
          "accountId": "632669837482532864",
          "content": "Did you see it? How hungry are you now? 😳\n\n#teen #petite #young #bigass #ass #pussy #cameltoe #naked #nude",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743339668,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "762714508241154048",
              "pos": 0,
              "contentType": 1,
              "contentId": "762684919011876864"
            },
            {
              "postId": "762714508241154048",
              "pos": 1,
              "contentType": 42001,
              "contentId": "762684906542211072"
            }
          ],
          "likeCount": 65,
          "replyCount": 7,
          "wallIds": [],
          "mediaLikeCount": 487,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "763799645498388480",
          "accountId": "355798632533860352",
          "content": "If you were here, Rina and I would fight for your dick 😏 @RinaTattoo \n\nWho would you fuck first? 😜",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743598385,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "763799645498388480",
              "pos": 0,
              "contentType": 1,
              "contentId": "763799644139433984"
            }
          ],
          "likeCount": 77,
          "replyCount": 11,
          "wallIds": [],
          "mediaLikeCount": 1167,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": [
            {
              "start": 58,
              "end": 68,
              "handle": "rinatattoo",
              "accountId": "721685520521895936"
            }
          ],
          "tipAmount": 5000
        },
        {
          "id": "764127047495725057",
          "accountId": "614203153939701762",
          "content": "my landlord caught me masturbating!!!!\n-\n#fyp #teen #asian #wasian #public #outdoors #young #petite #masturbation #fingering #small #stepsis #goth #alt #egirl #caught #free",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743676444,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "764127047495725057",
              "pos": 0,
              "contentType": 1,
              "contentId": "764127046149353473"
            },
            {
              "postId": "764127047495725057",
              "pos": 1,
              "contentType": 7100,
              "contentId": "763278560164065280"
            }
          ],
          "likeCount": 27,
          "replyCount": 2,
          "wallIds": [],
          "mediaLikeCount": 365,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "764979227228381185",
          "accountId": "763721773920296960",
          "content": "𝗔𝗡𝗔𝗟 𝗠𝗔𝗦𝗧𝗨𝗥𝗕𝗔𝗧𝗜𝗢𝗡 𝗪𝗜𝗧𝗛 𝗔𝗦𝗦𝗛𝗢𝗟𝗘 𝗚𝗔𝗣𝗘\n\nI take the dildo and use my hands to spread the lube on it. Then I lie down and insert the dildo completely in my ass and start jerking off 💦 \nI pull the dildo out a few times to show you how my asshole gaping ✨ \n\n𝗔𝗖𝗖𝗘𝗦𝗦 𝗪𝗜𝗧𝗛 «𝗩𝗜𝗣 👑» 𝗦𝗨𝗕𝗦𝗖𝗥𝗜𝗣𝗧𝗜𝗢𝗡 𝗢𝗥 𝗛𝗜𝗚𝗛𝗘𝗥 \n\n#anal #asshole #gape #schoolgirl #teen #young #18 #babyface #blonde #dildo #masturbation ",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743879619,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "764979227228381185",
              "pos": 0,
              "contentType": 1,
              "contentId": "764488855662370816"
            }
          ],
          "likeCount": 54,
          "replyCount": 6,
          "wallIds": [],
          "mediaLikeCount": 1190,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "765359166326841344",
          "accountId": "600098569193529344",
          "content": "A little car quickie 🙈🚗\n\n#fyp #lesbian #teen #petite #pussy #public #girlgirl #boobs #young #barelylegal #carsex ",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1743970204,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "765359166326841344",
              "pos": 0,
              "contentType": 1,
              "contentId": "765359164812701696"
            }
          ],
          "likeCount": 277,
          "replyCount": 23,
          "wallIds": [],
          "mediaLikeCount": 2456,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "767787714111086593",
          "accountId": "632669837482532864",
          "content": "Do you like hairy pussies? 😏\n\n#teen #petite #young #bigass #ass #pussy #cameltoe #naked #nude #hairy",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1744549215,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "767787714111086593",
              "pos": 0,
              "contentType": 1,
              "contentId": "766300267342737408"
            }
          ],
          "likeCount": 70,
          "replyCount": 9,
          "wallIds": [],
          "mediaLikeCount": 393,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "769405052346441728",
          "accountId": "717871284519710720",
          "content": "What am I going to do now????\n\n#teen #luck #slut #bop #college #luck #spinthewheel #prizes #college #collegeslut #collegedorm #teen #nude #young #petite #girl #hornyteen #hotvideo #fyp #bikini #18yearsold #nude #teennude #horny #collegestudent #young #petite #girl #hornyteen #hotvideo #fyp ",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1744934818,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "769405052346441728",
              "pos": 0,
              "contentType": 1,
              "contentId": "769359757239656448"
            }
          ],
          "likeCount": 60,
          "replyCount": 11,
          "wallIds": [],
          "mediaLikeCount": 1161,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "770747405791408128",
          "accountId": "635819933409751040",
          "content": "#petite #teen #young #barelylegal #boobs #pussy #pinkpussy #xsmall",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1745254860,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "770747405791408128",
              "pos": 0,
              "contentType": 1,
              "contentId": "770746921349296128"
            }
          ],
          "likeCount": 6,
          "wallIds": [],
          "mediaLikeCount": 628,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "771849562154348544",
          "accountId": "636133253325004801",
          "content": "I was a really good girl today! 😇 I play on console and with Daddy's big hard cock in MY PINK HOLE 😳 He stretched my lil pussy REALLY hard until I felt so much hot cum spurting into my tummy..🥺 I could feel it in my coochie it is very warm! 🤗 It feels so funny squeezing cum out of my used hole.. 🤭\n\nFull video here: https://fansly.com/post/767500961424875520\n\n#young #sex #creampied #cum #freeuse #creampie #breeding #little #petite #teen",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1745517635,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "771849562154348544",
              "pos": 0,
              "contentType": 1,
              "contentId": "771791413066080256"
            }
          ],
          "likeCount": 2,
          "wallIds": [],
          "mediaLikeCount": 176,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "773450856472256512",
          "accountId": "719990128717606913",
          "content": "watch my very first blowjob video ever - full video only on @tinyangelxxx for Extreme Tier subscribers <3<3\n\n#blowjob #facial #cuminmouth #cumslut #young #cute #babyface #deepthroat #sloppy ",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1745899413,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "773450856472256512",
              "pos": 0,
              "contentType": 1,
              "contentId": "773450852483473411"
            }
          ],
          "likeCount": 86,
          "replyCount": 12,
          "wallIds": [],
          "mediaLikeCount": 2351,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": [
            {
              "start": 60,
              "end": 72,
              "handle": "tinyangelxxx",
              "accountId": "394064703367688192"
            }
          ]
        },
        {
          "id": "781361720831516672",
          "accountId": "397844556042743808",
          "content": "My experimental outfit. I changed my mind about showing it, but my friend said it was beautiful. Do you like such a bright style?\n\nP.s: I liked this picture from my subscriber so much that I decided to add it.\n\n#bigass #bigtits  #petite #blonde #young #pussy #teen #tightpussy #nsfw #feet  #barelylegal #teengirl #boobs #schoolgirl #tiny #18  #naughty #gfe #explore #sexy #tiktok #girlnextdoor #hornyteen #college #cameltoe #lewd #egirl #tiny #twerk #foryoupage #fansly #creator #content #newcontent #subscribe #exclusive #gym #foryou #bdsm",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1747785510,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "781361720831516672",
              "pos": 0,
              "contentType": 1,
              "contentId": "781361720143654912"
            },
            {
              "postId": "781361720831516672",
              "pos": 1,
              "contentType": 1,
              "contentId": "781402354028126208"
            }
          ],
          "likeCount": 12,
          "replyCount": 3,
          "wallIds": [],
          "mediaLikeCount": 22,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "782057018276978691",
          "accountId": "738120971654799361",
          "content": "Are you a 🍒 or 🍑 guy? Be honest\n\n#ass #boobs #young #pussy #bigass #bigboobs #chubby #curvy #pawg #thick #sexting #belly #breeding #mombod #thighs #busty",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1747951282,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "782057018276978691",
              "pos": 0,
              "contentType": 1,
              "contentId": "781954107505127424"
            }
          ],
          "likeCount": 33,
          "postReplyPermissionFlags": [
            {
              "id": "782057018318921729",
              "postId": "782057018276978691",
              "type": 0,
              "flags": 2,
              "metadata": ""
            },
            {
              "id": "782057018318921730",
              "postId": "782057018276978691",
              "type": 0,
              "flags": 4,
              "metadata": "{}"
            }
          ],
          "replyCount": 5,
          "wallIds": [],
          "mediaLikeCount": 56,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "785439159551533056",
          "accountId": "736768099428081664",
          "content": "Lol🤭🤭\n\n#virgin#pussy#student#teen#smalltits#horny#cutie#young#petite#babygirl#barelylegal#mastrubation#nsfw#stepsis#gfe#18#18yo#younggirl#schoolgirl#shy#legalteen#school",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1748757647,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "785439159551533056",
              "pos": 0,
              "contentType": 1,
              "contentId": "785274273806692352"
            }
          ],
          "likeCount": 0,
          "wallIds": [],
          "mediaLikeCount": 36,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "785733438245187584",
          "accountId": "750438352355864576",
          "content": "My eyes say ‘innocent’… my hands say ‘guilty as charged’. 😇🔥\n#fyp #egirl #cute #alt #babyface #tattoo #petite #teen\n#bigtits #skinny #young #nude #tits #pussy #sexy #body\n#pussy #petite #horny #girl #tits #naked #nude\n#boobs #feet #TikTok #heels #tights #skirt\n#cells #stockings #schoolgirl #Fansly #FanslyGirl\n#FanslyModel #FanslyExclusive #Boobs #Titties\n#BigBoobs #Busty #BoobLover #Cleavage #NSFW\n#SpicyContent #Lingerie #Curvy #Thick #SexyVibes\n#Tease #FreeTease #HotAF #AltGirl #18plus #PPV\n#SubscribeNow #UnlockMe #WatchMe",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1748827809,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "785733438245187584",
              "pos": 0,
              "contentType": 1,
              "contentId": "785583601101058048"
            }
          ],
          "likeCount": 3,
          "wallIds": [],
          "mediaLikeCount": 7,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "786068611012767744",
          "accountId": "354231770675159040",
          "content": "⁠would you want my perfect soft body on top or under you? ❣️\n\n#fyp #horny #teen #tight #naked #smalltits #cute #petite #gfe #feet #socks #tits #nipples #pussy #wetpussy #young #barelylegal",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1748907720,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "786068611012767744",
              "pos": 0,
              "contentType": 1,
              "contentId": "785268485579087872"
            }
          ],
          "likeCount": 5,
          "wallIds": [],
          "mediaLikeCount": 5,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "787749306491744257",
          "accountId": "732609428909469696",
          "content": "I’m no angel😈 and you know it❤️‍🔥\n\n#fyp  #schoolgirl  #daddygirl #skinny #babyface #18 #teen #young #kinky #ass",
          "fypFlags": 2,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1749308429,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "787749306491744257",
              "pos": 0,
              "contentType": 1,
              "contentId": "787605303045664772"
            }
          ],
          "likeCount": 1,
          "wallIds": [],
          "mediaLikeCount": 3,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "787779688205393921",
          "accountId": "724022160137400320",
          "content": "Was that enough… or are you begging for more? 😜\n\n#petite #young #daddysgirl #babyface #tinytits #pussy #teen #tightpussy #ass #butt #nsfw #naked #feet #barelylegal #striptease #boobs #schoolgirl #tiny #naughty #gfe #explore #sexy #tiktok #girlnextdoor #cameltoe #lewd #egirl #twerk #foryoupage #fansly #creator #content #newcontent #subscribe #exclusive #gym #foryou",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1749315673,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "787779688205393921",
              "pos": 0,
              "contentType": 1,
              "contentId": "787759864788033536"
            }
          ],
          "likeCount": 9,
          "postReplyPermissionFlags": [
            {
              "id": "787779688251531264",
              "postId": "787779688205393921",
              "type": 0,
              "flags": 2,
              "metadata": ""
            }
          ],
          "replyCount": 1,
          "wallIds": [],
          "mediaLikeCount": 23,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": []
        },
        {
          "id": "788812257533370368",
          "accountId": "355798632533860352",
          "content": "@wollfans ⬅️ Follow me & receive a gift in BIO 🎁😈\n\n#tights #stockings #teen #young #petite #blonde #pussy #boobs #masturbation ",
          "fypFlags": 0,
          "inReplyTo": null,
          "inReplyToRoot": null,
          "replyPermissionFlags": null,
          "createdAt": 1749561857,
          "expiresAt": null,
          "attachments": [
            {
              "postId": "788812257533370368",
              "pos": 0,
              "contentType": 1,
              "contentId": "788812256589651968"
            }
          ],
          "likeCount": 0,
          "wallIds": [],
          "mediaLikeCount": 10,
          "totalTipAmount": 0,
          "attachmentTipAmount": 0,
          "accountMentions": [
            {
              "start": 0,
              "end": 8,
              "handle": "wollfans",
              "accountId": "355798632533860352"
            }
          ]
        }
      ],
      "tips": [],
      "tipGoals": [],
      "stories": []
    }
  }
}
```
