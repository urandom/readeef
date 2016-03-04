var ipc = require('electron').ipcRenderer;

var result = JSON.parse(ipc.sendSync('get-config-item-sync', 'show-notifications'));
if (result && result.success && result.value !== null && !result.value) {
	return;
}

window.addEventListener('updates-available', function(event) {
	var notification = new Notification('New articles', {
			'body': 'Feed "' + event.detail.title + '" has been updated',
	});

	notification.onclick = function() {
		ipc.send('focus-main-window');
	};
});
