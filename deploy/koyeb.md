# Despliegue de Whatomate para Petos en Koyeb

Koyeb debe ejecutar Whatomate como **un servicio web nuevo**. El dominio
`economic-inga-yesidcompany-2398b713.koyeb.app` pertenece al servicio
`notion-mcp-koyeb`; no se debe reemplazar para desplegar Whatomate.

## Servicios necesarios

1. PostgreSQL persistente. Crear una base de datos y un usuario propios para
   Whatomate, incluso si se reutiliza un servidor PostgreSQL existente.
2. Redis en un servicio privado accesible desde Whatomate. La aplicación no
   arranca sin conexión a Redis.
3. Servicio web `whatomate-petos`, conectado a este repositorio y a la rama
   elegida. Seleccionar el builder **Dockerfile** y la ruta
   `docker/Dockerfile`, exponer el puerto `8080` por HTTP y usar `/ready`
   como comprobación de salud.

El archivo `docker/docker-compose.yml` se usa únicamente para desarrollo local;
Koyeb no crea automáticamente sus servicios `db` y `redis` al construir el
Dockerfile.

## Variables del servicio web

Configurar las siguientes variables en Koyeb. Guardar contraseñas, claves y
tokens como **Secrets**, nunca en Git. El prefijo y el doble guion bajo son
obligatorios: `WHATOMATE_<SECCIÓN>__<CAMPO>`.

| Variable | Valor |
| --- | --- |
| `WHATOMATE_APP__ENVIRONMENT` | `production` |
| `WHATOMATE_APP__DEBUG` | `false` |
| `WHATOMATE_APP__ENCRYPTION_KEY` | Clave aleatoria de 32 caracteres o más; conservarla en futuros despliegues |
| `WHATOMATE_SERVER__ALLOWED_ORIGINS` | URL HTTPS exacta del servicio nuevo |
| `WHATOMATE_DATABASE__HOST` | Host privado de PostgreSQL |
| `WHATOMATE_DATABASE__PORT` | Puerto de PostgreSQL, normalmente `5432` |
| `WHATOMATE_DATABASE__USER` | Usuario propio de Whatomate |
| `WHATOMATE_DATABASE__PASSWORD` | Contraseña de ese usuario |
| `WHATOMATE_DATABASE__NAME` | Base de datos propia de Whatomate |
| `WHATOMATE_DATABASE__SSL_MODE` | Modo SSL exigido por PostgreSQL |
| `WHATOMATE_REDIS__HOST` | Host privado de Redis |
| `WHATOMATE_REDIS__PORT` | Puerto de Redis, normalmente `6379` |
| `WHATOMATE_REDIS__PASSWORD` | Contraseña de Redis, si aplica |
| `WHATOMATE_REDIS__TLS` | `true` si Redis exige TLS |
| `WHATOMATE_JWT__SECRET` | Clave aleatoria distinta, de 32 caracteres o más |
| `WHATOMATE_DEFAULT_ADMIN__EMAIL` | Correo del administrador inicial |
| `WHATOMATE_DEFAULT_ADMIN__PASSWORD` | Contraseña inicial única de 16 caracteres o más |
| `WHATOMATE_RATE_LIMIT__ENABLED` | `true` |
| `WHATOMATE_RATE_LIMIT__TRUST_PROXY` | `true` detrás del proxy de Koyeb |
| `WHATOMATE_WHATSAPP__APP_ID` | `1837938394295721` (app de Meta `petos`) |
| `WHATOMATE_WHATSAPP__APP_SECRET` | Clave secreta de esa app |
| `WHATOMATE_WHATSAPP__WEBHOOK_VERIFY_TOKEN` | Token aleatorio para verificar el webhook |

El arranque en producción rechaza la contraseña de administrador de ejemplo,
la clave JWT de ejemplo y la ausencia de clave de cifrado. En una base de datos
nueva, el administrador se crea una sola vez durante la migración.

## Meta y almacenamiento

Tras desplegar, configurar en Meta la URL
`https://<dominio-nuevo>.koyeb.app/api/webhook` y el mismo token de verificación.
El número comercial debe incorporarse al WhatsApp Business Platform por el
flujo apropiado antes de conectar una cuenta en Whatomate. Si el número sigue
en la app móvil de WhatsApp Business, utilizar el flujo de coexistencia para
conservar el uso en el teléfono.

El directorio local `/app/uploads` de un servicio web de Koyeb es efímero.
Para conservar archivos multimedia entre despliegues, configurar el backend
S3 de Whatomate o un almacenamiento persistente compatible antes de usarlo
con clientes reales.
