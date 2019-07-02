Newshound UI
===========

Written with AngularJS and a lil help from Yeoman.

Development requires node (npm), bower and grunt.

Install:

    npm install
    bower install

Build:

    grunt build  <--- start a concatenated and uglified build. results will go to 'dist'
    
Development:

    grunt serve <-- start a quick non-concatenated and non-uglified build for development. 
                    starts a server at '0.0.0.0:9000' with auto-refresh.
    grunt serve:dist <-- create a concatentated and uglified build
                    starts a server at '0.0.0.0:9000' with auto-refresh
