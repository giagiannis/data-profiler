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

