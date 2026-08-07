const CACHE_VERSION = 'wgm-v1';
const STATIC_CACHE = `${CACHE_VERSION}-static`;
const RUNTIME_CACHE = `${CACHE_VERSION}-runtime`;

const STATIC_ASSETS = [
  '/',
  '/login',
  '/favicon',
  '/static/dist/js/adminlte.min.js',
  '/static/dist/css/adminlte.min.css',
  '/static/plugins/jquery/jquery.min.js',
  '/static/plugins/bootstrap/js/bootstrap.bundle.min.js',
  '/static/plugins/select2/js/select2.full.min.js',
  '/static/plugins/jquery-validation/jquery.validate.min.js',
  '/static/plugins/toastr/toastr.min.css',
  '/static/plugins/toastr/toastr.min.js',
  '/static/plugins/fontawesome-free/css/all.min.css',
  '/static/plugins/icheck-bootstrap/icheck-bootstrap.min.css',
  '/static/plugins/select2/css/select2.min.css',
  '/static/plugins/jquery-tags-input/dist/jquery.tagsinput.min.js',
  '/static/plugins/jquery-tags-input/dist/jquery.tagsinput.min.css',
  '/static/custom/img/logo.png',
  '/static/custom/img/favicon-96x96.png',
  '/static/custom/img/favicon.svg',
  '/static/custom/img/apple-touch-icon.png',
  '/static/custom/img/web-app-manifest-192x192.png',
  '/static/custom/img/web-app-manifest-512x512.png',
  '/static/custom/img/site.webmanifest',
  '/static/custom/js/helper.js',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(STATIC_CACHE).then((cache) => {
      return cache.addAll(STATIC_ASSETS.map((url) => new Request(url, { cache: 'reload' })));
    }).then(() => {
      return self.skipWaiting();
    })
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => !name.startsWith(CACHE_VERSION))
          .map((name) => caches.delete(name))
      );
    }).then(() => {
      return self.clients.claim();
    })
  );
});

self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});

self.addEventListener('fetch', (event) => {
  const request = event.request;
  const url = new URL(request.url);

  if (request.method !== 'GET') {
    return;
  }

  if (url.origin !== self.location.origin) {
    return;
  }

  const isAPI = url.pathname.startsWith('/api/') || url.pathname.startsWith('/test-hash');
  const isStatic = url.pathname.startsWith('/static/');
  const isNavigation = request.mode === 'navigate';

  if (isAPI) {
    event.respondWith(networkFirst(request, RUNTIME_CACHE, 50));
    return;
  }

  if (isNavigation) {
    event.respondWith(networkFirst(request, RUNTIME_CACHE, 10));
    return;
  }

  if (isStatic) {
    event.respondWith(cacheFirst(request, STATIC_CACHE));
    return;
  }

  event.respondWith(staleWhileRevalidate(request, RUNTIME_CACHE));
});

async function cacheFirst(request, cacheName) {
  const cache = await caches.open(cacheName);
  const cached = await cache.match(request);
  if (cached) {
    return cached;
  }
  try {
    const response = await fetch(request);
    if (response.ok) {
      cache.put(request, response.clone());
    }
    return response;
  } catch (err) {
    return cached || new Response('Offline', { status: 503, statusText: 'Offline' });
  }
}

async function networkFirst(request, cacheName, maxEntries) {
  const cache = await caches.open(cacheName);
  try {
    const response = await fetch(request);
    if (response.ok) {
      cache.put(request, response.clone());
      trimCache(cache, maxEntries);
    }
    return response;
  } catch (err) {
    const cached = await cache.match(request);
    if (cached) {
      return cached;
    }
    if (request.mode === 'navigate') {
      const staticCache = await caches.open(STATIC_CACHE);
      const fallback = await staticCache.match('/');
      if (fallback) return fallback;
    }
    return new Response('Offline', { status: 503, statusText: 'Offline' });
  }
}

async function staleWhileRevalidate(request, cacheName) {
  const cache = await caches.open(cacheName);
  const cached = await cache.match(request);
  const fetchPromise = fetch(request).then((response) => {
    if (response.ok) {
      cache.put(request, response.clone());
    }
    return response;
  }).catch(() => cached);
  return cached || fetchPromise;
}

async function trimCache(cache, maxEntries) {
  if (!maxEntries) return;
  const keys = await cache.keys();
  if (keys.length > maxEntries) {
    for (let i = 0; i < keys.length - maxEntries; i++) {
      cache.delete(keys[i]);
    }
  }
}
