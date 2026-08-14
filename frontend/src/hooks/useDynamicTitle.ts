import { useEffect } from 'react';
import { useRouterState } from '@tanstack/react-router';

export function useDynamicTitle() {
  const matches = useRouterState({ select: (s) => s.matches });

  useEffect(() => {
    const lastMatch = matches[matches.length - 1];
    if (lastMatch) {
      const path = lastMatch.routeId;
      
      let title = 'POS Fiplex';
      if (path.includes('login')) {
        title = 'Login | POS Fiplex';
      } else if (path.includes('_dashboard')) {
        let page = path.split('/').pop();
        if (page === '' || page === undefined || page === '_dashboard') {
          page = 'Dashboard';
        } else {
          page = page.charAt(0).toUpperCase() + page.slice(1).replace('-', ' ');
        }
        title = `${page} | POS Fiplex`;
      }
      
      document.title = title;
    }
  }, [matches]);
}
