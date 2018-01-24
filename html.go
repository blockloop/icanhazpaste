package main

import "html/template"

const HTMLForm = `
<!DOCTYPE html>
<html>
<body>

<form action="/" method="post">
  First name:</br>
  <textarea name="clip" style="width: 75%; height: 300px;"></textarea>
  </br>
  <input type="submit" value="Submit">
  </br>
</form>

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
