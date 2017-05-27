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

function createHeatMap(url, orderBy) {
		function findMinElement(data) {
				var minElement = 1.0;
				for(i=0;i<data.length;i++) {
						if (data[i].value < minElement) {
								minElement = data[i].value
						}
				}
				return minElement
		}

		function sortElements(x_elements, orderBy, data) {
				if (orderBy != "" ) {
						var idx = x_elements.indexOf(orderBy);
						var tosort = []
						for(i=0;i<data.length;i++) {
								if (data[i].x == orderBy) {
										tosort.push({label:data[i].y, value:data[i].value})
								}
						}
						tosort.sort(function(a,b) { return  - (a.value - b.value) });
						return tosort.map(function (d) {return d.label});
						//			x_elements = tosort.map(function (d) {return d.label});
						//			y_elements = x_elements;
				}	
				return x_elements
		}


		d3.csv(url, function ( response ) {
				var data = response.map(function( item ) {
						var newItem = {};
						newItem.x= item.x;
						newItem.y= item.y;
						newItem.value = item.value;
						return newItem;
				})

				var x_elements = d3.set(data.map(function( item ) { return item.x; } )).values(),
						y_elements = d3.set(data.map(function( item ) { return item.y; } )).values();
				minElement = findMinElement(data);

				x_elements = sortElements(x_elements, orderBy, data)
				y_elements = x_elements
				var xScale = d3.scale.ordinal()
						.domain(x_elements)
						.rangeBands([0, x_elements.length * itemSize]);

				var xAxis = d3.svg.axis()
						.scale(xScale)
				//        .tickFormat(function (d) {
				//          return d;
				//    })
						.ticks(10)
						.orient("top");

				var yScale = d3.scale.ordinal()
						.domain(y_elements)
						.rangeBands([0, y_elements.length * itemSize]);

				var yAxis = d3.svg.axis()
						.scale(yScale)
						.tickFormat(function (d) {
								return d;
						})
						.orient("left");


				domainValues =[]
				step=0.1
				if (minElement>0) {
						step = minElement/7.0;
				}
				var x=parseFloat(minElement)
				while(x<1.0){
						domainValues.push(x)
						x = x+step;
				}
				domainValues.push(1.01)

				cols = ["#fff7fb","#ece7f2","#d0d1e6","#a6bddb","#74a9cf","#3690c0","#0570b0","#045a8d","#023858", "#001432"].reverse()
				var colorScale = d3.scale.threshold()
						.domain(domainValues)
						.range(cols);

				var svg = d3.select('.heatmap')
						.append("svg")
				//		.attr("width", "auto")
				//		.attr("height", "auto")
						.attr("width", width + margin.left + margin.right)
						.attr("height", height + margin.top + margin.bottom)
						.append("g")
						.attr("transform", "translate(" + margin.left + "," + margin.top + ")");

				var tooltip = d3.select("body").append("div")
						.attr("class", "tooltip")			
						.style("opacity", 0);

				var cells = svg.selectAll('rect')
						.data(data)
						.enter().append('g').append('rect')
						.attr('class', 'cell')
						.attr('width', cellSize)
						.attr('height', cellSize)
						.attr('y', function(d) { return yScale(d.y); })
						.attr('x', function(d) { return xScale(d.x); })
						.attr('yor', function(d) { return d.y; })
						.attr('xor', function(d) { return d.x; })
						.attr('fill', function(d) { return colorScale(d.value); })
						.attr('value', function(d) { return d.value; })
						.on('mouseover', function() {
								d3.select(this)
										.style('fill', 'white');
						})
						.on('mouseout', function() {
								d3.select(this)
										.style('fill', '');
						})
						.on('click', function() {
								var obj = d3.select(this);
								tooltip.html(obj.attr("xor")+", "+obj.attr("yor")+":"+obj.attr("value"))
										.style("left", (d3.event.pageX) + "px")
										.style("top", (d3.event.pageY - 50) + "px");

								if (tooltip.style("opacity") == "0") {
										tooltip.style("opacity",1.0);
								} else {
										tooltip.style("opacity",0.0);
								}
						})
						.style("stroke", '#555');

				svg.append("g")
						.attr("class", "y axis")
						.call(yAxis)
						.selectAll('text')
						.attr('font-weight', 'normal')
						.attr('font-size', '8px')
						.on('click', function(d){
								// TODO: I do a better transition here
								//			sorted = sortElements(x_elements, d, data);
								//			var t = svg.transition().duration(100);
								//			t.selectAll(".cell").attr("y", function(d) {return sorted.indexOf(d.y- 1)*cellSize; })
								$(".heatmap").remove();
								$("#container").append("<div class='heatmap'></div>");
								createHeatMap(smID, d);
						});

				svg.append("g")
						.attr("class", "x axis")
						.call(xAxis)
						.selectAll('text')
						.attr('font-weight', 'normal')
						.attr('font-size', '8px')
						.style("text-anchor", "start")
						.attr("dx", ".8em")
						.attr("dy", ".5em")
						.attr("transform", function (d) {
								return "rotate(-65)";
						});
		});
}

function create3DScatterPlot(coordinatesID, labelsID, targetDiv) {
		var lines= $( "#"+coordinatesID ).html().split("\n");
		var labels= $( "#"+labelsID ).html().split("\n");
		var data = [];
		for(var i=0;i<lines.length;i++) {
				str = lines[i].split(",");
				var tuple = {};
				var defineTuple = false;
				if (str.length > 0){
						tuple.x = parseFloat(str[0]);
						if (!isNaN(tuple.x)){
								defineTuple = true;
						}
				} else {
						tuple.x = 0.0;
				}
				if (str.length > 1){
						tuple.y = parseFloat(str[1]);
						defineTuple = true;
				} else {
						tuple.y = 0.0;
				}
				if (str.length > 2){
						tuple.z = parseFloat(str[2]);
						defineTuple = true;
				} else {
						tuple.z = 0.0;
				}
				if (defineTuple) {
						tuple.name = labels[i];
						data.push(tuple);
				}
		}
		var maxElem={x:"",y:"",z:""}, minElem = {x:"", y:"", z:""};
		for(var i=0;i<data.length;i++) {
				o = data[i]
				if (maxElem.x == "" || maxElem.x < o.x) {
						maxElem.x = o.x;
				}
				if (maxElem.y == "" || maxElem.y < o.y) {
						maxElem.y = o.y;
				}
				if (maxElem.z == "" || maxElem.z < o.z) {
						maxElem.z = o.z;
				}
				if (minElem.x == "" || minElem.x > o.x) {
						minElem.x = o.x;
				}
				if (minElem.y == "" || minElem.y > o.y) {
						minElem.y = o.y;
				}
				if (minElem.z == "" || minElem.z > o.z) {
						minElem.z = o.z;
				}

		}
		// set tuples color
		for(var i=0;i<data.length;i++) {
				var r="0", g="0", b="0";
				maxCol = 150;
				if (maxElem.x > 0 ) {
					r = parseInt(Math.floor((data[i].x - minElem.x)/(maxElem.x - minElem.x) * maxCol))
				}
				if (maxElem.y > 0 ) {
				g = parseInt(Math.floor((data[i].y - minElem.y)/(maxElem.y - minElem.y) * maxCol))
				}
				if (maxElem.z > 0 ) {
				b = parseInt(Math.floor((data[i].z - minElem.z)/(maxElem.z - minElem.z) * maxCol))
				}
				data[i].color = "rgb("+r+","+g+","+b+")";
				console.log(data[i].color)
				
		}
		// Give the points a 3D feel by adding a radial gradient
		Highcharts.getOptions().colors = $.map(Highcharts.getOptions().colors, function (color) {
				return {
						radialGradient: {
								cx: 0.4,
								cy: 0.3,
								r: 0.5
						},
						stops: [
								[0, color],
								[1, Highcharts.Color(color).brighten(-0.2).get('rgb')]
						]
				};
		});

		// Set up the chart
		var chart = new Highcharts.Chart({
				chart: {
						renderTo: targetDiv,
						margin: 100,
						type: 'scatter',
						options3d: {
								enabled: true,
								alpha: 10,
								beta: 30,
								depth: 800,
								viewDistance: 5,
								fitToPlot: false,
								frame: {
										bottom: { size: 1, color: 'rgba(0,0,0,0.02)' },
										back: { size: 1, color: 'rgba(0,0,0,0.04)' },
										side: { size: 1, color: 'rgba(0,0,0,0.06)' }
								}
						},
				},
				title: {
						text: 'Dataset space'
				},
				subtitle: {
						text: 'Click and drag the plot area to rotate in space'
				},
				tooltip: {
						//pointFormat: '<span style=color:{point.color}>\u25CF</span> {series.name}: <b>{point.y}</b><br/>'
						pointFormatter: function(){ return this.name+"<br/>("+parseFloat(this.x)+", "+parseFloat(this.y)+", "+parseFloat(this.z)+")";}
				},
				plotOptions: {
						scatter: {
								width: 10,
								height: 10,
								depth: 10
						}
				},
				yAxis: {
						title: { 
								text : "PC2"
						}
				},
				xAxis: {
						title: { 
								text : "PC1"
						},
						gridLineWidth: 1
				},
				zAxis: {
						title: { 
								text : "PC3"
						},
						showFirstLabel: false
				},
				legend: {
						enabled: false
				},
				series: [{
						name: 'Dataset',
						colorByPoint: false,
						data : data
				}]
		});


		// Add mouse events for rotation
		$(chart.container).on('mousedown.hc touchstart.hc', function (eStart) {
				eStart = chart.pointer.normalize(eStart);

				var posX = eStart.pageX,
						posY = eStart.pageY,
						alpha = chart.options.chart.options3d.alpha,
						beta = chart.options.chart.options3d.beta,
						newAlpha,
						newBeta,
						sensitivity = 6; // lower is more sensitive

				$(document).on({
						'mousemove.hc touchdrag.hc': function (e) {
								// Run beta
								newBeta = beta + (posX - e.pageX) / sensitivity;
								chart.options.chart.options3d.beta = newBeta;

								// Run alpha
								newAlpha = alpha + (e.pageY - posY) / sensitivity;
								chart.options.chart.options3d.alpha = newAlpha;

								chart.redraw(false);
						},
						'mouseup touchend': function () {
								$(document).off('.hc');
						}
				});
		});
};

