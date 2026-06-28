import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '@/api.ts'

const routesForAuthedOnly: Record<string, boolean> = {
  "incident-detail": true,
  "incidents": true,
  "incidents-new": true,
}

const routesForUnauthedOnly: Record<string, boolean> = {
  "log-in": true,
}

const publicRoutes: Record<string, boolean> = {
  "trial-accounts": true,
  "sandbox": true,
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/entry',
      name: 'entry',
      component: () => { },
    },
    {
      path: '/incident-detail/:id',
      name: 'incident-detail',
      component: () => import('@/views/IncidentDetailView.vue'),
    },
    {
      path: '/log-in',
      name: 'log-in',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/incidents',
      name: 'incidents',
      component: () => import('@/views/IncidentsView.vue'),
    },
    {
      path: '/incidents/new',
      name: 'incidents-new',
      component: () => import('@/views/IncidentCreateView.vue'),
    },
    {
      path: '/sandbox',
      name: 'sandbox',
      component: () => import('@/views/SandboxView.vue'),
    },
    {
      path: '/registration',
      name: 'registration',
      component: () => import('@/views/RegistrationView.vue'),
    },
    {
      path: '/trial-accounts',
      name: 'trial-accounts',
      component: () => import('@/views/TrialAccounts.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'default',
      redirect: "/entry",
    },
  ],
})

router.beforeEach(async (to) => {
  if (typeof to.name === "string") {

    const authed = await isAuthenticated()
    if (to.name == "entry") {
      return authed == true ? { name: "incidents" } : { name: "log-in" }
    }
    if (authed && routesForUnauthedOnly[to.name]) {
      return { name: "incidents" }
    }
    if (!authed && routesForAuthedOnly[to.name]) {
      return { name: "log-in" }
    }
  }
})

export default router
