# Configuración básica
IMAGE_NAME ?= dperez9/cert-manager-webhook-duckdns
IMAGE_TAG ?= latest
BIN_NAME := webhook
OUT := $(CURDIR)/_out

# Crear carpeta _out si no existe
$(shell mkdir -p "$(OUT)")

# Variables internas para Helm
CHART_NAME := cert-manager-webhook-duckdns
CHART_PATH := deploy/$(CHART_NAME)

# 1. Compilar el binario
build:
	go mod tidy
	go build -v -o $(OUT)/$(BIN_NAME) main.go

# 2. Ejecutar tests locales
# test:
# 	go test -v ./...

# 3. Crear imagen Docker
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# 4. Subir imagen
docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

# 5. Renderizar manifiesto desde Helm Chart
render:
	helm template \
		$(CHART_NAME) $(CHART_PATH) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) > $(OUT)/rendered-manifest.yaml

# 6. Limpieza
clean:
	rm -rf $(OUT)

.PHONY: build test docker-build docker-push render clean
