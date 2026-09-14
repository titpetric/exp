{{define "userCard"}}
<div class="user-card">
<img src="{{.Avatar}}" alt="{{.Name}} avatar" />
<h2>{{.Name}}</h2>
<a href="{{.Link}}">Profile</a>
</div>
{{end}}

{{template "userCard" .}}
