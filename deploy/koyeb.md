# Despliegue de Whatomate para Petos en Koyeb

El dominio `economic-inga-yesidcompany-2398b713.koyeb.app` pertenece al
servicio `notion-mcp-koyeb`. Su propietario autorizó reemplazarlo por
Whatomate, conservando ese dominio. Antes de cambiar su código o variables,
preparar PostgreSQL, Redis y los secretos, y comprobar que el coste estimado
del servicio web sea **$0**. La instancia actual `eNano` cuesta $1,61/mes.

Koyeb permite una sola instancia web Free por organización. En esta cuenta la
ocupa `charvis30`, así que hay que liberar esa plaza antes de cambiar
`notion-mcp-koyeb` a Free. El propietario autorizó liberar `charvis30`.

## Servicios necesarios

1. PostgreSQL persistente en **Neon Free**, con base de datos propia para
   Whatomate. La base Free de Koyeb solo permite 5 horas activas al mes, por
   lo que no se utilizará para este número comercial.
2. **Upstash Redis Free**, conectado por el protocolo Redis/TCP con TLS. La
   aplicación no arranca sin conexión a Redis. Mantener el consumo dentro de
   las cuotas gratuitas de Upstash.
3. Servicio web existente `notion-mcp-koyeb`, conectado a este repositorio y
   a `main`. Seleccionar la instancia **Free**, builder **Dockerfile** y ruta
   `docker/Dockerfile`, exponer el puerto `8080` por HTTP y usar `/ready`
   como comprobación de salud. Free duerme tras una hora sin tráfico y se
   reactiva con una nueva petición; tener en cuenta el arranque en frío para
   los webhooks de Meta. El servidor también ejecuta un worker de campañas,
   pero ese worker se detiene mientras la instancia está dormida. No usar
   campañas programadas que dependan de ejecución continua en este plan.

El archivo `docker/docker-compose.yml` se usa únicamente para desarrollo local;
Koyeb no crea automáticamente sus servicios `db` y `redis` al construir el
Dockerfile.

## Variables del servicio web

Configurar las siguientes variables en Koyeb. Guardar contraseñas, claves y
tokens como **Secrets**, nunca en Git ni como valores literales de variables.
Referenciarlos como `{{ secret.NOMBRE }}` en el editor de variables; Koyeb
puede incluir los valores literales en el enlace de despliegue generado por
su interfaz. El prefijo y el doble guion bajo son
obligatorios: `WHATOMATE_<SECCIÓN>__<CAMPO>`.

| Variable | Valor |
| --- | --- |
| `WHATOMATE_APP__ENVIRONMENT` | `production` |
| `WHATOMATE_APP__DEBUG` | `false` |
| `WHATOMATE_APP__ENCRYPTION_KEY` | Clave aleatoria de 32 caracteres o más; conservarla en futuros despliegues |
| `WHATOMATE_SERVER__ALLOWED_ORIGINS` | `https://economic-inga-yesidcompany-2398b713.koyeb.app` |
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
| `WHATOMATE_WHATSAPP__CONFIG_ID` | `1444149661001102` (configuración `Petos WhatsApp`) |
| `WHATOMATE_WHATSAPP__WEBHOOK_VERIFY_TOKEN` | Token aleatorio para verificar el webhook |

El arranque en producción rechaza la contraseña de administrador de ejemplo,
la clave JWT de ejemplo y la ausencia de clave de cifrado. En una base de datos
nueva, el administrador se crea una sola vez durante la migración.

## Meta y almacenamiento

Tras desplegar, configurar en Meta la URL
`https://economic-inga-yesidcompany-2398b713.koyeb.app/api/webhook` y el
mismo token de verificación.
El número comercial debe incorporarse al WhatsApp Business Platform por el
flujo apropiado antes de conectar una cuenta en Whatomate. Si el número sigue
en la app móvil de WhatsApp Business, utilizar el flujo de coexistencia para
conservar el uso en el teléfono.

El directorio local `/app/uploads` de un servicio web de Koyeb es efímero.
Para conservar archivos multimedia entre despliegues, configurar el backend
S3 de Whatomate o un almacenamiento persistente compatible antes de usarlo
con clientes reales.
