package com.dodecaneser.rastreador.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dodecaneser.rastreador.ui.theme.CyberColors

enum class TacticalTab(val title: String, val badge: String) {
    RADAR_MAP("RADAR / MAPA", "GEO"),
    WARDRIVING("WARDRIVING RF", "802.11"),
    INTEL("INTEL RED", "BGP/IP-ID")
}

@Composable
fun TacticalBottomNavigationBar(
    selectedTab: TacticalTab,
    onTabSelected: (TacticalTab) -> Unit,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(CyberColors.SurfaceDark)
            .navigationBarsPadding()
            .padding(horizontal = 8.dp, vertical = 6.dp),
        horizontalArrangement = Arrangement.SpaceAround,
        verticalAlignment = Alignment.CenterVertically
    ) {
        TacticalTab.values().forEach { tab ->
            val isSelected = tab == selectedTab
            val backgroundColor = if (isSelected) CyberColors.SurfaceElevated else Color.Transparent
            val textColor = if (isSelected) CyberColors.CyberCyan else CyberColors.TextMuted

            Box(
                modifier = Modifier
                    .weight(1f)
                    .clip(RoundedCornerShape(6.dp))
                    .background(backgroundColor)
                    .clickable { onTabSelected(tab) }
                    .padding(vertical = 8.dp, horizontal = 4.dp),
                contentAlignment = Alignment.Center
            ) {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = "[ ${tab.badge} ]",
                        style = MaterialTheme.typography.labelSmall.copy(
                            fontSize = 9.sp,
                            color = if (isSelected) CyberColors.MatrixGreen else CyberColors.TextMuted
                        )
                    )
                    Spacer(modifier = Modifier.height(2.dp))
                    Text(
                        text = tab.title,
                        style = MaterialTheme.typography.labelMedium.copy(
                            color = textColor,
                            fontSize = 11.sp
                        )
                    )
                }
            }
        }
    }
}
