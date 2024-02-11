#!/bin/bash

mkdir /tmp/go-build

# Read the list of username@ip_address
echo "Enter the list of username@ip_address, separated by commas:"
read -r input_list

# Convert the input list into an array
IFS=',' read -r -a user_ip_list <<< "$input_list"

# Default SSH password
SSH_PASS="123456"

# Check for existing SSH keys
ssh_key="$HOME/.ssh/id_rsa"
if [ ! -f "$ssh_key" ]; then
    echo "SSH key not found, generating one..."
    ssh-keygen -t rsa -b 4096 -f "$ssh_key" -N ""
fi

# Iterate over the user@ip list
for user_ip in "${user_ip_list[@]}"; do
    echo "Processing $user_ip..."

    # Install sshpass if not installed
    if ! command -v sshpass &> /dev/null; then
        echo "sshpass not found, installing..."
        sudo apt-get install -y sshpass
    fi

    # Copy SSH key to remote machine using sshpass
    sshpass -p "$SSH_PASS" ssh-copy-id -i "$ssh_key.pub" "$user_ip"

    # SSH into each machine and modify ~/.bashrc, create directories
    sshpass -p "$SSH_PASS" ssh -oStrictHostKeyChecking=no "$user_ip" << 'EOF'
#    echo "export GOTMPDIR=/tmp/go-build" >> ~/.bashrc
#    mkdir -p /tmp/go-build
#    mkdir -p ~/Projects
#    mkdir -p ~/Projects/configs
#    wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
#    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
#    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
EOF
    # Upload the files to the remote machine
    sshpass -p "$SSH_PASS" scp configs/worker.yaml "$user_ip":~/Projects/configs/worker.yaml
    sshpass -p "$SSH_PASS" scp worker "$user_ip":~/Projects/
    sshpass -p "$SSH_PASS" scp worker "$user_ip":~/
done

echo "Script execution completed."
