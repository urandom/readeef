var Menu = require('electron').Menu;
var menuTemplate = require('./menus/' + process.platform + '.json');

exports.init = function(readeef) {
	function bindMenuItems(menuItems) {
		menuItems.forEach(function bindMenuItemFn (menuItem) {
			// If there is a role, continue
			if (menuItem.role !== undefined) {
				return;
			}

			// If there is a separator, continue
			if (menuItem.type === 'separator') {
				return;
			}

			// If there is a submenu, recurse it
			if (menuItem.submenu) {
				bindMenuItems(menuItem.submenu);
				return;
			}

			// Otherwise, find the function for our command
			var cmd = menuItem.command;
			if (cmd === 'application:about') {
				menuItem.click = readeef.openAboutWindow;
			} else if (cmd === 'application:show-settings') {
				menuItem.click = readeef.openConfigWindow;
			} else if (cmd === 'application:quit') {
				menuItem.click = readeef.quitApplication;
			} else if (cmd === 'window:reload') {
				menuItem.click = readeef.reloadWindow;
			} else if (cmd === 'window:toggle-dev-tools') {
				menuItem.click = readeef.toggleDevTools;
			} else if (cmd === 'window:toggle-full-screen') {
				menuItem.click = readeef.toggleFullScreen;
			} else {
				throw new Error('Could not find function for menu command "' + cmd + '" ' +
					'under label "' + menuItem.label + '"');
			}
		});
	}
	bindMenuItems(menuTemplate.menu);
	Menu.setApplicationMenu(Menu.buildFromTemplate(menuTemplate.menu));
}
