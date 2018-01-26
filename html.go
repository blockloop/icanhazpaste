package main

import "html/template"

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
