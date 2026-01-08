#!/bin/bash

# Actualizar los paquetes e instalar Git
sudo apt update
sudo apt install -y git

# Configurar nombre de usuario y correo electrónico de Git
echo "Configurando Git..."
read -p "Introduce tu nombre de usuario de Git: " git_username
read -p "Introduce tu correo electrónico de Git: " git_email
git config --global user.name "$git_username"
git config --global user.email "$git_email"

# Generar una clave SSH
echo "Generando clave SSH..."
ssh-keygen -t ed25519 -C "$git_email" -f ~/.ssh/id_ed25519 -N ""

# Añadir la clave SSH al agente SSH
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519

# Mostrar la clave SSH y pedir que se añada a GitHub
echo "Clave SSH generada. Añádela a GitHub en https://github.com/settings/keys"
cat ~/.ssh/id_ed25519.pub
echo "Abriendo GitHub en el navegador..."
xdg-open "https://github.com/settings/keys"

# Probar la conexión SSH a GitHub
echo "Probando conexión SSH con GitHub..."
ssh -T git@github.com

echo "Configuración completada. Puedes clonar repositorios usando SSH."
