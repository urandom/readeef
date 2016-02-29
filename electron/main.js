var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.
var storage = require('./storage')
var checker = require('./url-checker')

// Report crashes to our server.
require('crash-reporter').start({
		companyName: "Sugr",
		submitURL: "https://github.com/urandom/readeef/issues",
});

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
var mainWindow = null;

// Quit when all windows are closed.
app.on('window-all-closed', function() {
	// On OS X it is common for applications and their menu bar
	// to stay active until the user quits explicitly with Cmd + Q
	if (process.platform != 'darwin') {
		app.quit();
	}
});

// This method will be called when Electron has finished
// initialization and is ready to create browser windows.
app.on('ready', function() {
	var lastWindowState = storage.get("lastWindowState");
	if (lastWindowState === null) {
		lastWindowState = {
			width: 1024,
			height: 768,
			maximized: false 
		} 
	}

	// Create the browser window.
	mainWindow = new BrowserWindow({
		x: lastWindowState.x,
		y: lastWindowState.y,
		width: lastWindowState.width, 
		height: lastWindowState.height
	});

	// and load the index.html of the app.
	if (storage.get("url")) {
		mainWindow.loadURL(storage.get('url'));
	} else {
		mainWindow.loadURL('file://' + __dirname + '/index.html');
	}


	// Open the DevTools.
	// mainWindow.openDevTools();
	mainWindow.on('close', function() {
		var bounds = mainWindow.getBounds(); 
		storage.set("lastWindowState", {
			x: bounds.x,
			y: bounds.y,
			width: bounds.width,
			height: bounds.height,
			maximized: mainWindow.isMaximized()
		});
	});

	// Emitted when the window is closed.
	mainWindow.on('closed', function() {
		// Dereference the window object, usually you would store windows
		// in an array if your app supports multi windows, this is the time
		// when you should delete the corresponding element.
		mainWindow = null;
	});
});
