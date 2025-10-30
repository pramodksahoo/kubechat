import { Navigate, createRootRoute, createRoute, createRouter } from '@tanstack/react-router';
import { kcDetailsSearch, kcListSearch } from '@/types';

import FourOFourError from "@/components/app/Errors/404Error";
import GenericError from "@/components/app/Errors/GenericError";
import { KubeConfiguration } from '@/components/app/KubeConfiguration';
import { KubeChat } from '@/KubeChat';
import { KcDetails } from '@/components/app/Common/Details';
import { KcList } from '@/components/app/Common/List';

const AppWrapper = ({ component }: { component: JSX.Element }) => {
  return (
    component
  );
};

const rootRoute = createRootRoute({
  component: () => <KubeChat />
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => <Navigate to="/kcconfig" />
});

const kcList = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$config/list',
  component: () => <AppWrapper component={<KcList />} />,
  validateSearch: (search: Record<string, unknown>): kcListSearch => {
    return {
      cluster: String(search.cluster) || '',
      resourcekind: String(search.resourcekind) || '',
      ...(search.group ? {group: String(search.group)}: {}),
      ...(search.kind ? {kind: String(search.kind)}: {}),
      ...(search.resource ? {resource: String(search.resource)}: {}),
      ...(search.version ? {version: String(search.version)}: {}),
      ...(search.plan ? {plan: String(search.plan)} : {}),
    };
  }
});

const kcDetails = createRoute({
  getParentRoute: () => rootRoute,
  path: '/$config/details',
  component: () => <AppWrapper component={<KcDetails />} />,
  validateSearch: (search: Record<string, unknown>): kcDetailsSearch => ({
    cluster: String(search.cluster) || '',
    resourcekind: String(search.resourcekind) || '',
    resourcename: String(search.resourcename) || '',
    group: search.group ? String(search.group) : '',
    kind: search.kind? String(search.kind) : '',
    resource: search.resource ? String(search.resource) : '',
    version:search.version ? String(search.version) : '',
    namespace: search.namespace ? String(search.namespace) : '',
    ...(search.plan ? { plan: String(search.plan) } : {}),
  })
});



const kubeConfigurationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/kcconfig',
  component: KubeConfiguration,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  kubeConfigurationRoute,
  kcList,
  kcDetails
]);

const router = createRouter({
  routeTree,
  defaultNotFoundComponent: () => <FourOFourError />,
  defaultErrorComponent: () => <GenericError />,
  defaultPreload: 'intent',
  defaultStaleTime: 5000,
});

// Register things for typesafety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

export {
  router,
  kcList,
  kcDetails
};
