function createSM(id) {
		$.get("/datasets/"+id+"/newsm", function(data) {
		$( "<div>").html(data).dialog();
		});
}
