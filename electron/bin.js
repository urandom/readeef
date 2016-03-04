#!/usr/bin/env node
var spawn = require('child_process').spawn;
var electronPath = require('electron-prebuilt');

var args = [__dirname];

args = args.concat(process.argv.slice(2));

// Run electron on our application and forward all stdio
spawn(electronPath, args, {stdio: [0, 1, 2]});
