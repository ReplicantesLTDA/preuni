import { useFonts as useExpoFonts } from 'expo-font';
import { CaveatBrush_400Regular } from '@expo-google-fonts/caveat-brush';
import { PatrickHand_400Regular } from '@expo-google-fonts/patrick-hand';
import { ArchitectsDaughter_400Regular } from '@expo-google-fonts/architects-daughter';
import { Kalam_400Regular, Kalam_700Bold } from '@expo-google-fonts/kalam';
import { Caveat_400Regular, Caveat_700Bold } from '@expo-google-fonts/caveat';

export function useAppFonts(): { fontsReady: boolean; error: Error | null } {
  const [loaded, error] = useExpoFonts({
    CaveatBrush: CaveatBrush_400Regular,
    PatrickHand: PatrickHand_400Regular,
    ArchitectsDaughter: ArchitectsDaughter_400Regular,
    Kalam: Kalam_400Regular,
    KalamBold: Kalam_700Bold,
    Caveat: Caveat_400Regular,
    CaveatBold: Caveat_700Bold,
  });
  return { fontsReady: loaded, error: error ?? null };
}
