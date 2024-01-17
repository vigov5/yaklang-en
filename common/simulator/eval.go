// Package simulator
// @Author bcy2007  2023/8/23 15:32
package simulator

const observer = `
()=>{
	const config = { attributes: true, childList: true, subtree: true, characterData: true };
	window.added = "";
	window.addednode = null;
	// Callback function executed when changes are observed
	const callback = function(mutationsList, observer) {
		// Use traditional 'for loops' for IE 11
		for(let mutation of mutationsList) {
			if (mutation.type === 'childList') {
				//window.node = node;
				for (let node of mutation.addedNodes) {
					addednode = node
					// added += node.innerHTML;
					if (node.innerHTML !== undefined) {
						added += node.innerHTML
					} else if (node.data !== undefined){
						added += node.data
					} else {
						added += node.nodeValue
					} 
				}
			}
			else if (mutation.type === 'attributes') {
			}
			else if (mutation.type === 'characterData') {
				added += mutation.target.data;
			}
		}
	};
	// Create an observer instance and pass in the callback function
	window.observer = new MutationObserver(callback);
	// Start observing the target node with the above configuration
	observer.observe(document, config);
}
`

const getObverserResult = `
()=>{
	if (typeof(added) !== "string"){
		return ""
	}
	try {
		ahrefs = addednode.getElementsByTagName("a")
		if (ahrefs.length !== 0){
			ahrefs[0].click()
		}
	} catch (err) {}
	observer.disconnect();
	return added;
}
`
