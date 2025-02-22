import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "smsgatewayapi/sms-gateway",
    },
    {
      type: "category",
      label: "phone",
      items: [
        {
          type: "doc",
          id: "smsgatewayapi/register-phone",
          label: "Register new phone",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "smsgatewayapi/update-fcm-token",
          label: "updateFcmToken",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "smsgatewayapi/get-phone-by-id",
          label: "getPhoneById",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "account",
      items: [
        {
          type: "doc",
          id: "smsgatewayapi/register-account",
          label: "Register new account",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "smsgatewayapi/get-account-by-id",
          label: "getAccountById",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "sms",
      items: [
        {
          type: "doc",
          id: "smsgatewayapi/get-sms-by-id",
          label: "Get an sms",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "smsgatewayapi/get-messages",
          label: "Get all filtered sms",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "smsgatewayapi/send-sms",
          label: "Send a new sms",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "webhooks",
      items: [
        {
          type: "doc",
          id: "smsgatewayapi/enable-webhook-delivery",
          label: "Enable webhook delivery",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "reports",
      items: [
        {
          type: "doc",
          id: "smsgatewayapi/report-message-status",
          label: "reportMessageStatus",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
