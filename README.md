## RUN
```text
go run ./cmd/server
```

* servidor siempre activo
* primer paso: instalar pm2
```bash
npm install -g pm2
```
* segundo paso:
```bash
pm2 start go --name similia-backend -- run ./cmd/server
```

## GIT

* …or create a new repository on the command line
```text
echo "# similia_be" >> README.md
git init
git add README.md
git commit -m "first commit"
git branch -M main
git remote add origin https://github.com/zeron-team/similia_be.git
git push -u origin main
```
* …or push an existing repository from the command line
```text
git remote add origin https://github.com/zeron-team/similia_be.git
git branch -M main
git push -u origin main
```

## 🧰 Comandos útiles para backend Go con pm2

Acción                          Comando
Ver logs.                  pm2 logs similia-backend
Reiniciar backend.         pm2 restart similia-backend
Detener backend            pm2 stop similia-backend
Eliminar del gestor        pm2 delete similia-backend
Ver detalles del proceso   pm2 describe similia-backend
Ver errores puntuales.     pm2 monit

