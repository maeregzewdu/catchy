import { createRouter, createWebHistory } from 'vue-router'
import { useAccountStore } from '@/stores/accounts'

const HomeView = () => import('@/views/HomeView.vue')
const TrapInboxView = () => import('@/views/TrapInboxView.vue')
const AccountFolderView = () => import('@/views/AccountFolderView.vue')
const SearchView = () => import('@/views/SearchView.vue')
const SettingsView = () => import('@/views/SettingsView.vue')

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: HomeView },
    { path: '/dashboard', redirect: '/dashboard/trap' },
    {
      path: '/dashboard/trap',
      component: TrapInboxView,
    },
    {
      path: '/dashboard/trap/:messageId',
      component: TrapInboxView,
      props: true,
    },
    {
      path: '/dashboard/accounts/:accountId/:folder',
      component: AccountFolderView,
      props: true,
    },
    {
      path: '/dashboard/accounts/:accountId/:folder/:messageId',
      component: AccountFolderView,
      props: true,
    },
    {
      path: '/dashboard/search',
      component: SearchView,
    },
    {
      path: '/dashboard/search/:messageId',
      component: SearchView,
      props: true,
    },
    {
      path: '/dashboard/settings',
      component: SettingsView,
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach(async () => {
  const accountStore = useAccountStore()
  if (accountStore.accounts.length === 0 && !accountStore.loading) {
    await accountStore.fetchAccounts()
  }
})

export default router
