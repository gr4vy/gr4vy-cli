export const decodeJWT = (token: string): Record<string, any> => {
  const [header, claims] = token.split(".");

  return {
    header: JSON.parse(Buffer.from(header, "base64").toString()),
    claims: JSON.parse(Buffer.from(claims, "base64").toString()),
  };
};
