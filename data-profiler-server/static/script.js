function createPopup(url, dialogTitle) {
$.get( url, function(data) {
	var d= $( "<div>").html(data).dialog({
			width:"auto", 
			height:"auto", title: dialogTitle});
});
}

function toggleVisibility(objId) {
	if ($( objId ).is(":visible")) {
			$( objId).hide();
	} else {
			$( objId).show();
	}
}
