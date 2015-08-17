var gulp = require('gulp');
var compass = require('gulp-compass');

gulp.task('compass', function() {
  gulp.src('./src/*.scss')
    .pipe(compass({
      config_file: './config.rb',
      css: 'build',
      sass: 'sass'
    }))
    .pipe(gulp.dest('build'));
});

gulp.task('default', ['compass']);
