import { renderHook } from '@testing-library/react-native';

const mockUseExpoFonts = jest.fn();
jest.mock('expo-font', () => ({
  useFonts: (...args: unknown[]) => mockUseExpoFonts(...args),
}));

import { useAppFonts } from '@/theme/useFonts';

describe('useAppFonts', () => {
  it('reports fontsReady once expo-font finishes loading', () => {
    mockUseExpoFonts.mockReturnValue([true, undefined]);
    const { result } = renderHook(() => useAppFonts());
    expect(result.current).toEqual({ fontsReady: true, error: null });
  });

  it('surfaces a load error and fontsReady=false', () => {
    const err = new Error('font load failed');
    mockUseExpoFonts.mockReturnValue([false, err]);
    const { result } = renderHook(() => useAppFonts());
    expect(result.current).toEqual({ fontsReady: false, error: err });
  });
});
