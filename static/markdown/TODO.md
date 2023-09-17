### OreCast tasks

- Ask for some data, and put it to s3, then test fetching this data and streaming.
- Add mongodb to meatada
- Add Google map to main page with site icon, the sites info should come from metadata which should supply geo locations
- Add storage endpoint to create bucket and upload data.
- Switch to restful endpoints, eg /storage/Cornell/bucket, add http delete, put, post methods to it
- Move code from frontend storage handler to data management service. Then, storage handler will call data management apis.
- Add stie registration form page with captcha.
- Add user registration form page with captcha. Decide where to keep users data. I think we need yet another service for that.
- Split site menu to submenu: registration, access, storage. The former will lead to site registration page. The second to current sites endpoint, and latter to site s3.
- Add data menu with management, access, viewer submenus. In management page we need form to create datasets , upload files, etc
- Create new repo for orecast client .
```
Orecast site add ...
Orecast site ls ...
Orecast meta add....
Orecast DBS add...
Orecast data create Cornell/bucket
Orecast data upload Cornell/bucket file
Orecast discover Cornell
```
- In metadata and discover repositories create handlers.go module.
- Add mongodb.go to common repo that we can use in other services.
- Add DBS codebase with simple schema.
- Add common/data where I should store all structures used in different modules to avoid code duplication.
