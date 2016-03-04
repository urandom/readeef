var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.
var ipcMain = require('electron').ipcMain;
var storage = require('./storage');
var checker = require('./url-checker');
var menu = require('./menu');
var pkg = require('./package.json');

// Report crashes to our server.
require('crash-reporter').start({
		companyName: "Sugr",
		submitURL: "https://github.com/urandom/readeef/issues",
});

var readeef = {
	// Keep a global reference of the window object, if you don't, the window will
	// be closed automatically when the JavaScript object is garbage collected.
	mainWindow: null,

	openAboutWindow: function() {
		var info = [
			// https://github.com/corysimmons/typographic/blob/2.9.3/scss/typographic.scss#L34
			'<div style="text-align: center; font-family: \'Helvetica Neue\', \'Helvetica\', \'Arial\', \'sans-serif\'">',
			'<h1>readeef</h1>',
			'<p>',
			'Version: ' + pkg.version,
			'<br/>',
			'Electron version: ' + process.versions.electron,
			'<br/>',
			'Node.js version: ' + process.versions.node,
			'<br/>',
			'Chromium version: ' + process.versions.chrome,
			'</p>',
			'</div>'
		].join('');
		var aboutWindow = new BrowserWindow({
				height: 180,
				//icon: assets['icon-32'],
				width: 400
		});
		aboutWindow.loadURL('data:text/html,' + info);
	},
	openConfigWindow: function () {
		var configWindow = new BrowserWindow({
				height: 440,
				//icon: assets['icon-32'],
				width: 620
		});
		configWindow.loadURL('file://' + __dirname + '/index.html');
	},
	quitApplication: function () {
		app.quit();
	},
	reloadWindow: function () {
		BrowserWindow.getFocusedWindow().reload();
	},
	toggleDevTools: function () {
		BrowserWindow.getFocusedWindow().toggleDevTools();
	},
	toggleFullScreen: function () {
		var focusedWindow = BrowserWindow.getFocusedWindow();
		// Move to other full screen state (e.g. true -> false)
		var wasFullScreen = focusedWindow.isFullScreen();
		var toggledFullScreen = !wasFullScreen;
		focusedWindow.setFullScreen(toggledFullScreen);
	},
};

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
	var windowOptions = {
		x: lastWindowState.x,
		y: lastWindowState.y,
		width: lastWindowState.width, 
		height: lastWindowState.height
	};

	// and load the index.html of the app.
	if (storage.get("url")) {
		windowOptions['node-integration'] = false;
		windowOptions['preload'] =  __dirname + '/browser.js';
		readeef.mainWindow = new BrowserWindow(windowOptions);
		readeef.mainWindow.loadURL(storage.get('url'));
	} else {
		readeef.mainWindow = new BrowserWindow(windowOptions);
		readeef.mainWindow.loadURL('file://' + __dirname + '/index.html?initial');
	}


	// Open the DevTools.
	// mainWindow.openDevTools();
	readeef.mainWindow.on('close', function() {
		var bounds = readeef.mainWindow.getBounds(); 
		storage.set("lastWindowState", {
			x: bounds.x,
			y: bounds.y,
			width: bounds.width,
			height: bounds.height,
			maximized: readeef.mainWindow.isMaximized()
		});
	});

	// Emitted when the window is closed.
	readeef.mainWindow.on('closed', function() {
		// Dereference the window object, usually you would store windows
		// in an array if your app supports multi windows, this is the time
		// when you should delete the corresponding element.
		readeef.mainWindow = null;
	});

	menu.init(readeef);

	ipcMain.on('focus-main-window', function(evt) {
		readeef.mainWindow.focus();
	});

	ipcMain.on('reload-main-window', function(evt) {
		readeef.mainWindow.reload();
	});
});
