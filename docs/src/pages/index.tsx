import type {ReactNode} from 'react';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

import { Redirect } from '@docusaurus/router';


export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Documentation for SMS Gateway">
      <main>
        <Redirect to="smsgatewayapi/sms-gateway" />;
      </main>
    </Layout>
  );
}
