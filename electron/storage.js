var app = require('app');
var fs = require('fs');
var path = require('path');
var ipcMain = require('electron').ipcMain;
var data = null;
var dataFilePath = path.join(app.getPath('userData'), 'data.json'); 

function load() {
	if (data !== null) {
		return;
	} 

	if (!fs.existsSync(dataFilePath)) {
		data = {};
		return;
	}

	data = JSON.parse(fs.readFileSync(dataFilePath, 'utf-8')); 
}

function save() {
	fs.writeFileSync(dataFilePath, JSON.stringify(data)); 
}

exports.set = function (key, value) {
	load();
	data[key] = value; 
	save();
}

exports.get = function (key) { 
	load();
	var value = null;
	if (key in data) {
		value = data[key];
	} 
	return value;
}

exports.unset = function (key) { 
	load();
	if (key in data) {
		delete data[key];
		save();
	} 
}

ipcMain.on('set-config-item-sync', function(evt, key, val) {
	exports.set(key, val);
	evt.returnValue = JSON.stringify({success: true});
});

ipcMain.on('unset-config-item-sync', function(evt, key) {
	exports.unset(key);
	evt.returnValue = JSON.stringify({success: true});
});

ipcMain.on('get-config-item-sync', function(evt, key) {
	var val = exports.get(key);
	evt.returnValue = JSON.stringify({success: true, value: val});
});
