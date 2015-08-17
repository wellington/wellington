To use this, you must symlink wt to compass.

`ln -s $(which wt) ${PATH}/compass`

Check that `compass -v` is pointed at wellington, then the following should work.

```
npm i gulp-compass --save-dev
gulp
```
