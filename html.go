package main

import "html/template"

const HTMLForm = `
<!DOCTYPE html>
<html>
<body>

<hgroup>
<h1>pbpaste</h1>
<h3>paste text and share</h1>
</hgroup>

<form action="/" method="post">
  Paste:</br>
  <textarea name="clip" style="width: 75%; height: 300px;"></textarea>
  </br>
  <input type="submit" value="Submit">
  </br>
</form>

<span>All pastes expire in 72 Hours</span>
</br>

</body>
</html>
`

var HTMLCreatedTemplate = template.Must(template.New("result").Parse(`
<!DOCTYPE html>
<html>
<body>

<script type="text/javascript">
	window.setTimeout(function(){
		window.location.replace("{{ .URL }}");
	},0);
</script>

</body>
</html>
`))
