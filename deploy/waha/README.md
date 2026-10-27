# Canal temporal WAHA para Petos

Este canal permite enviar **mensajes de texto** desde el chat y la API de Whatomate con un número alterno. No modifica la app de Meta ni usa el número terminado en **8226**. El backend verifica el teléfono asociado a WAHA antes de cada envío y rechaza ese número como emisor o destinatario.

## Servicio local

El servicio se ejecuta en Docker y solo escucha en `127.0.0.1:3000`. Sus credenciales están en `deploy/waha/.env` (ignorado por Git) y la sesión se conserva en el volumen `waha_waha-sessions`.

```sh
docker compose -f deploy/waha/compose.yaml up -d
```

Escribir el número alterno en `PETOS_WAHA_PHONE=` dentro del `.env` privado de esta carpeta. Abrir `http://127.0.0.1:3000/dashboard`, crear una sesión llamada `petos-temporal` y vincular **únicamente ese número** desde Dispositivos vinculados. No escanear el QR con el 8226.

## Conectar el backend

El backend de Koyeb necesita alcanzar WAHA por una URL HTTPS estable. Un túnel seguro desde este Mac es posible, pero debe permanecer encendido; `127.0.0.1` en Koyeb apunta a Koyeb y no a este Mac. No publicar el puerto 3000 directamente. WAHA recomienda usar API key y HTTPS en todo acceso remoto. Si se usa ngrok, ejecutar `ngrok http 3000 --traffic-policy-file deploy/waha/ngrok-policy.yaml` desde la raíz del repositorio; esa política expone solo los dos endpoints necesarios y WAHA sigue validando `X-Api-Key`. Una URL aleatoria de ngrok cambia al reiniciar el túnel; conviene usar un dominio estático gratuito de la cuenta o actualizar la variable en Koyeb tras cada reinicio.

Variables del backend (secretos en el proveedor de despliegue, nunca en Git):

```text
PETOS_WAHA_URL=https://URL-HTTPS-DEL-TUNEL
PETOS_WAHA_API_KEY=<mismo valor de WAHA_API_KEY>
PETOS_WAHA_SESSION=petos-temporal
PETOS_WAHA_PHONE=<numero alterno completo, sin + ni espacios>
```

Si el backend y WAHA corren en la misma máquina, `PETOS_WAHA_URL=http://127.0.0.1:3000` también funciona. El backend rechaza HTTP remoto. Si no se definen estas variables, el canal WAHA queda desactivado y Meta sigue funcionando.

En **Ajustes → Cuentas**, comprobar que WAHA indica “conectado” y pulsar **WAHA temporal**. Esta cuenta se añade sin reemplazar la cuenta de Meta como predeterminada. Para enviar desde Petos, seleccionar `WAHA temporal` en el chat o llamar al endpoint habitual:

```http
POST /api/contacts/{id}/messages
Content-Type: application/json

{"type":"text","content":{"body":"Hola"},"whatsapp_account":"WAHA temporal"}
```

La cuenta temporal no admite plantillas de Meta, archivos, botones, campañas ni llamadas. La recepción de mensajes por webhook WAHA aún no forma parte de este canal. Para retirarlo, eliminar la cuenta temporal de Whatomate y quitar las variables `PETOS_WAHA_*`; la integración oficial continúa disponible.

WAHA es un cliente no oficial y WhatsApp puede restringir cuentas que incumplan sus reglas de mensajería. Usar solo conversaciones esperadas por los destinatarios y revisar las [directrices de WhatsApp](https://www.whatsapp.com/legal/messaging-guidelines?lang=es). Referencias técnicas: [WAHA Send Text](https://waha.devlike.pro/docs/how-to/send-messages/), [sesiones](https://waha.devlike.pro/docs/how-to/sessions/) y [seguridad](https://waha.devlike.pro/docs/how-to/security/).
