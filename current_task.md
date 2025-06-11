Let's plan out a refactor for this app

1. @backend/src/db/schema.ts

- remove `tags.description`, `tags.flags`, `tags.isTracked`, `tag_history.recordedAt`
- remove any code related to those fields

2. Frontend

- Add a sidebar
  - Add a toolbox icon as the project logo
  - add a icon for "Tags"
  - add a worked status indicator at the bottom of the sidebar that just indicates that any worker is running
  - add a theme toggle button at the bottom of the sidebar
- main page
  - just a welcome page, explain the project and what it does
- tags page
  - add a table with the following columns:
    - Tag
    - View Count
    - Change (based on user selection of date range, use shadcn daterange picker with presets for 1 day, 1 week, 1 month, 3 months, 6 months, 1 year)
    - Last Updated
  - the table should be sortable, filterable and searchable (debounced)
  - the table should have a pagination
  - clicking on a tag should collapse the table and show the tag historical data with a date picker to select the date range
  - "Add Tag" button at the top right of the page
    - open a dialog with a text input and a button to add the tag
    - the tag should be added to the database and the worker should start tracking it
    - display a warning if the tag is already in the database and use that as the search query

3. Backend

- Remove the entire worker router
- Add 1 route that returns the status of the workers
  - running if any worker is running
  - idle if no worker is running
  - failed if any worker has failed
