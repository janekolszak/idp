#!/bin/bash
# Run in host to generate certificates for development
openssl req -nodes -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days XXX -subj '/CN=localhost'
