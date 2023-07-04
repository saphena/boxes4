# How does Boxes4 differ from earlier versions

There are major improvements in several areas: IT environment, usability, functionality, aesthetics, reliability.

## Reliability
The application has been refined to ensure that data is held and updated in a reliable fashion. Data entered is cleaned of unwanted whitespace and forced to uppercase where necessary. Searches now find poorly entered data. Errors are positively trapped and handled correctly rather than being allowed to fail silently.

## Usability
All interfaces now behave consistently and intuitively, including paging for all datasets, across all modern browsers and devices. 

All output now complies with web accessibility standards.

Paged results can be navigated with arrow keys, page up/down or swipe left/right. In update mode, all forms now load with autofocus in the appropriate field.

Result listings can be reordered on any column including counts. Pagesize can be varied by the user.

Because the application now consists of a single executable and a single diskfile it can be easily moved or copied to other computers for backup purposes or offline data access.

## Functionality
Searches now include all fields as well as partial matches.
Searches may be limited to particular fields such as Client or Owner.
All fields now editable directly. Editable fields shown in italics as visual clue.
Now able to filter search results by excluding old (review date) records.
Data export in JSON format as well as CSV is included.

## Aesthetics
Revised cleaner colourways and font choices combined with consistent styling thoughout make this easier on the eye.
Colour theme is selectable for individual users.

## IT environment
A single binary executable with no dependencies makes for a much simpler IT environment. The application can be easily installed and operated from cloud-based servers as well as laptops or tablets. No reliance on PHP. Can be configured to use any port.

The GO source code of the application, needed only for development, not operations, is now organised into a coherent, consistent and highly maintainable structure within a single repository.

Can be easily generated for Windows, Linux or Mac environments.