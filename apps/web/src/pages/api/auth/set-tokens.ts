// API route to set secure httpOnly cookies for JWT tokens
import type { NextApiRequest, NextApiResponse } from 'next';
import { serialize } from 'cookie';

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  const { accessToken, refreshToken } = req.body;

  if (!accessToken) {
    return res.status(400).json({ error: 'Access token is required' });
  }

  try {
    // Set secure httpOnly cookie for access token
    const accessTokenCookie = serialize('kubechat_token', accessToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'strict',
      path: '/',
      maxAge: 24 * 60 * 60, // 24 hours in seconds
    });

    const cookies = [accessTokenCookie];

    // Set refresh token cookie if provided
    if (refreshToken) {
      const refreshTokenCookie = serialize('kubechat_refresh_token', refreshToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        path: '/',
        maxAge: 7 * 24 * 60 * 60, // 7 days in seconds
      });
      cookies.push(refreshTokenCookie);
    }

    res.setHeader('Set-Cookie', cookies);
    res.status(200).json({ success: true });
  } catch (error) {
    console.error('Error setting secure tokens:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
}