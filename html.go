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

var HTMLHelpTemplate = template.Must(template.New("help").Parse(`
<!DOCTYPE html>
<html>
<head>
</head>
  <div>
   <pre>
   {{ .Text }}
   </pre>
  </div>
 <div>
 This software is open source at <a href="https://git.blockloop.io/blockloop/icanhazpaste">https://git.blockloop.io/blockloop/icanhazpaste</a>
 </div>

<script type="text/javascript"> var _paq = _paq || []; _paq.push(['trackPageView']); _paq.push(['enableLinkTracking']); (function() { var u="//scitylana.dokku.blockloop.io/"; _paq.push(['setTrackerUrl', u+'piwik.php']); _paq.push(['setSiteId', '1']); var d=document, g=d.createElement('script'), s=d.getElementsByTagName('script')[0]; g.type='text/javascript'; g.async=true; g.defer=true; g.src=u+'piwik.js'; s.parentNode.insertBefore(g,s); })(); </script>

</html>
`))
