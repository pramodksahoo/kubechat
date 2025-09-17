// API route to clear secure httpOnly cookies
import type { NextApiRequest, NextApiResponse } from 'next';
import { serialize } from 'cookie';

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    // Clear both access token and refresh token cookies
    const clearAccessTokenCookie = serialize('kubechat_token', '', {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      path: '/',
      expires: new Date(0), // Set expiry to past date to clear cookie
    });

    const clearRefreshTokenCookie = serialize('kubechat_refresh_token', '', {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      path: '/',
      expires: new Date(0), // Set expiry to past date to clear cookie
    });

    res.setHeader('Set-Cookie', [clearAccessTokenCookie, clearRefreshTokenCookie]);
    res.status(200).json({ success: true });
  } catch (error) {
    console.error('Error clearing secure tokens:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}