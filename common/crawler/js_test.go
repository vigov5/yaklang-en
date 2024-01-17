package crawler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMUSTPASS_JSHandle(t *testing.T) {
	var count = 0
	code := `console.log('1.js'); var deepUrl = 'deep.js';;
console.log('2.js'); fetch(deepUrl, {
	method: 'POST',
	headers: { 'HackedJS': "AAA"},
});


fetch('/misc/response/fetch/basic.action')
  .then(response => {
    if (!response.ok) {
      throw new Error('Network response was not ok ' + response.statusText);
    }
    return response.text();
  })
  .then(data => {
    console.log(data); // Here is your page content
  })
  .catch(error => {
    console.error('There has been a problem with your fetch operation:', error);
  });


;
// Create a new XMLHttpRequest object
var xhr = new XMLHttpRequest;

// Configure the request type as POST, and the target URL
xhr.open('POST', 'deep.js', true);

// Set required HTTP request headers
xhr.setRequestHeader('HackedJS', 'AAA');

// Set the callback function after the request is completed
xhr.onreadystatechange = function() {
  // Check whether the request is completed
  if (xhr.readyState === XMLHttpRequest.DONE) {
    // Check whether the request is successful
    if (xhr.status === 200) {
      // The request is successful, process the response data
      console.log(xhr.responseText);
    } else {
      // request fails, print status code
      console.error('Request failed with status:', xhr.status);
    }
  }
};

// Send the request, you can send it here Any required data
xhr.send();;
`
	HandleJS(false, []byte(`GET / HTTP/1.1
Host: www.example.com

`), code, func(b bool, bytes []byte) {
		count++
	})
	assert.Equal(t, 3, count)
}
