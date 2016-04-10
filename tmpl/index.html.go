<!DOCTYPE html>
<html>
<head>
	<title>DocBay</title>
</head>
<body>
	<h1>Doc Bay</h1>

	<h2>Projects</h2>
	<ul>
		{{range $key, $val := .ProjectVersions}}
		<li>{{$key}}
			<ul>
				{{range $val}}
				<li><a href="/{{$key}}/{{.}}/">{{.}}</a></li>
				{{end}}
			</ul>
		</li>
		{{end}}
	</ul>

	<h2>Upload Docs</h2>
	<form enctype="multipart/form-data" action="/_/upload" method="post">
		<label for="project">Project</label>
		<input type="text" name="project" id="project" />
		<br/>
		<label for="version">Version</label>
		<input type="text" name="version" id="version" />
		<br/>
		<label for="content">Content</label>
		<input type="file" name="content" />
		<br/>
		<input type="submit" value="Upload" />
	</form>
</body>
</html>
